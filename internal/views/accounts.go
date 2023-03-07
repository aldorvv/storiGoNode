package views

import (
	"database/sql"
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/storiGoNode/internal/aws"
	"github.com/storiGoNode/internal/db"
	"github.com/storiGoNode/internal/db/models"
	"github.com/storiGoNode/internal/mailing"
)

var REPO *sql.DB = db.GetRepo()

// List endpoint for `/accounts`
func GetAccounts(c *gin.Context) {
	const query = "SELECT * FROM accounts;"

	result, err := REPO.Query(query)
	var accounts []models.Account

	if err != nil {
		panic(err.Error())
	}

	for result.Next() {
		var account models.Account
		err = result.Scan(&account.ID, &account.File)
		if err != nil {
			panic(err.Error())
		}
		accounts = append(accounts, account)
	}
	c.IndentedJSON(http.StatusOK, accounts)
}

// Get endpoint for `/accounts/:id`
func GetAccount(c *gin.Context) {
	const query = "SELECT * FROM accounts WHERE id=?;"
	id := c.Param("id")
	var account models.Account

	result, err := REPO.Query(query, id)

	if err != nil {
		panic(err.Error())
	}

	for result.Next() {

		err = result.Scan(&account.ID, &account.File)
		if err != nil {
			panic(err.Error())
		}
	}
	c.IndentedJSON(http.StatusOK, account)
}

// Post endpoint for `/accounts/`
func PostAccount(c *gin.Context) {
	var newAccount models.Account
	const query = "INSERT INTO accounts (`file`) VALUES (?);"
	stmt, err := REPO.Prepare(query)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	defer stmt.Close()

	file, err := c.FormFile("file")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	newAccount.File = uploadFile(file)

	result, err := stmt.Exec(newAccount.File)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	accountID, _ := result.LastInsertId()
	newAccount.ID = int(accountID)

	csvFile, _ := file.Open()
	err = createMovesFromFile(csvFile, newAccount)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newAccount)

}

// Get endpoint for `/accounts/:id/summary?email`
func GetSummary(c *gin.Context) {
	id := c.Param("id")
	recipient := c.Query("email")

	if !doAccountExist(id) {
		c.IndentedJSON(http.StatusNotFound, "Account not found.")
		return
	}

	stats := getStatsFor(id)

	templateData := mailing.TemplateData{
		Total:       strconv.FormatFloat(stats["total"], 'g', 5, 64),
		AvgCredit:   strconv.FormatFloat(stats["avgCred"], 'g', 5, 64),
		AvgDebit:    strconv.FormatFloat(stats["avgDeb"], 'g', 5, 64),
		MonthCounts: getMonthlyCounts(id),
	}

	m := mailing.NewMail("Your summary is ready 🚀", "summary.html", templateData)
	m.SendTo(recipient)

	c.IndentedJSON(http.StatusOK, "Email sent to "+recipient+"!")
}

// Create moves for each new account from parsed file
func createMovesFromFile(csvFile multipart.File, belongsTo models.Account) error {
	const query = "INSERT INTO move (`kind`, `quantity`, `date`, `accountID`) VALUES (?, ?, ?, ?);"

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	defer csvFile.Close()

	for _, row := range records[1:] {
		var kind int
		transaction, err := strconv.ParseFloat(row[2], 64)

		if err != nil {
			return err
		}

		if transaction < 0 {
			kind = -1 // Debit
		} else {
			kind = 1 // Credit
		}

		date, err := time.Parse("2006/1/2", "2023/"+row[1])
		if err != nil {
			return err
		}

		move := models.Move{
			AccountID: belongsTo.ID,
			Kind:      kind,
			Quantity:  transaction,
			Date:      date,
		}

		stmt, _ := REPO.Prepare(query)
		_, err = stmt.Exec(move.Kind, move.Quantity, move.Date, move.AccountID)

		if err != nil {
			return err
		}
	}
	return nil
}

// Upload file to an S3 bucket
func uploadFile(file *multipart.FileHeader) string {
	scv := aws.GetConnection()
	return scv.UploadFile(file)
}

func getAccountMoves(accountID string) []models.Move {
	const query = "SELECT * FROM move WHERE accountID=?;"
	result, err := REPO.Query(query, accountID)

	if err != nil {
		panic(err.Error())
	}

	var moves []models.Move

	for result.Next() {
		var move models.Move
		err = result.Scan(&move.ID, &move.Kind, &move.Quantity, &move.Date, &move.AccountID)
		if err != nil {
			panic(err.Error())
		}
		moves = append(moves, move)
	}
	return moves
}

// Verifies an account is present in the database
func doAccountExist(accountID string) bool {
	const query = "SELECT COUNT(*) FROM accounts WHERE id=?;"
	var accountExists int
	REPO.QueryRow(query, accountID).Scan(&accountExists)

	return accountExists != 0
}

// Returns a map with the calculated stats for an account
func getStatsFor(accountID string) map[string]float64 {
	const query = "SELECT COUNT(*) FROM move WHERE kind=? AND accountID=?;"
	var debSum, credSum, total float64

	for _, move := range getAccountMoves(accountID) {
		if move.Kind == -1 {
			debSum += move.Quantity
		} else {
			credSum += move.Quantity
		}
		total += move.Quantity
	}

	var debCount int
	REPO.QueryRow(query, -1, accountID).Scan(&debCount)
	avgDeb := debSum / float64(debCount)

	var credCount int
	REPO.QueryRow(query, 1, accountID).Scan(&credCount)
	avgCred := credSum / float64(credCount)

	stats := make(map[string]float64)
	stats["avgDeb"] = avgDeb
	stats["avgCred"] = avgCred
	stats["total"] = total

	return stats
}

// Return how many moves the user did in each month, ignore zeros
func getMonthlyCounts(accountID string) []mailing.MonthCount {
	const query = "SELECT COUNT(*) FROM move WHERE MONTH(date)=? AND accountID=?;"
	monthCounts := []mailing.MonthCount{}

	for i := 1; i <= 12; i++ {
		currentCount := 0
		REPO.QueryRow(query, i, accountID).Scan(&currentCount)
		if currentCount > 0 {
			monthCounts = append(
				monthCounts,
				mailing.MonthCount{
					Month: time.Month(i).String(),
					Count: currentCount,
				},
			)
		}
	}

	return monthCounts
}
