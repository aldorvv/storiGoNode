package views

import (
	"encoding/csv"
	"fmt"
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

func GetAccounts(c *gin.Context) {
	const query = "SELECT * FROM accounts;"
	repo := db.GetRepo()

	result, err := repo.Query(query)
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

func GetAccount(c *gin.Context) {
	const query = "SELECT * FROM accounts WHERE id=?;"
	repo := db.GetRepo()
	id := c.Param("id")
	var account models.Account

	result, err := repo.Query(query, id)

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

func PostAccount(c *gin.Context) {
	var newAccount models.Account
	repo := db.GetRepo()
	const query = "INSERT INTO accounts (`file`) VALUES (?);"
	stmt, err := repo.Prepare(query)
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
	err = insertMovesFromFile(csvFile, newAccount)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newAccount)

}

func GetSummary(c *gin.Context) {
	repo := db.GetRepo()
	id := c.Param("id")

	var accountExists int
	repo.QueryRow("SELECT COUNT(*) FROM accounts WHERE id=?;", id).Scan(&accountExists)
	if accountExists == 0 {
		c.IndentedJSON(http.StatusNotFound, "Account not found.")
		return
	}

	const query = "SELECT * FROM move WHERE accountID=?;"
	recipient := c.Query("email")
	result, err := repo.Query(query, id)

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

	var debSum, credSum, total float64

	for _, move := range moves {
		if move.Kind == -1 {
			debSum += move.Quantity
		} else {
			credSum += move.Quantity
		}
		total += move.Quantity
	}

	var debCount int
	repo.QueryRow("SELECT COUNT(*) FROM move WHERE kind=-1 AND accountID=?;", id).Scan(&debCount)
	avgDeb := debSum / float64(debCount)
	fmt.Println(debCount)

	var credCount int
	repo.QueryRow("SELECT COUNT(*) FROM move WHERE kind=1 AND accountID=?;", id).Scan(&credCount)
	avgCred := credSum / float64(credCount)
	fmt.Println(credCount)

	type MonthCount struct {
		Month string
		Count int
	}

	monthCounts := []MonthCount{}

	for i := 1; i <= 12; i++ {
		currentCount := 0
		repo.QueryRow("SELECT COUNT(*) FROM move WHERE MONTH(date)=? AND accountID=?;", i, id).Scan(&currentCount)
		if currentCount > 0 {
			monthCounts = append(monthCounts, MonthCount{Month: time.Month(i).String(), Count: currentCount})
		}
	}

	templateData := struct {
		Total       string
		AvgCredit   string
		AvgDebit    string
		MonthCounts []MonthCount
	}{
		Total:       strconv.FormatFloat(total, 'g', 5, 64),
		AvgCredit:   strconv.FormatFloat(avgCred, 'g', 5, 64),
		AvgDebit:    strconv.FormatFloat(avgDeb, 'g', 5, 64),
		MonthCounts: monthCounts,
	}

	m := mailing.NewMail("Your summary is ready 🚀", "summary.html", templateData)
	m.SendTo(recipient)

	c.IndentedJSON(http.StatusOK, "Email sent to "+recipient+"!")
}

func insertMovesFromFile(csvFile multipart.File, belongsTo models.Account) error {
	const query = "INSERT INTO move (`kind`, `quantity`, `date`, `accountID`) VALUES (?, ?, ?, ?);"
	repo := db.GetRepo()

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

		stmt, _ := repo.Prepare(query)
		_, err = stmt.Exec(move.Kind, move.Quantity, move.Date, move.AccountID)

		if err != nil {
			return err
		}
	}
	return nil
}

func uploadFile(file *multipart.FileHeader) string {
	scv := aws.GetConnection()
	return scv.UploadFile(file)
}
