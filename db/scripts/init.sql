CREATE DATABASE stori;
USE stori;

CREATE TABLE accounts (
    id INT AUTO_INCREMENT NOT NULL,
    `file` VARCHAR(255) NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE move (
    id INT AUTO_INCREMENT NOT NULL,
    kind INT NOT NULL,
    quantity FLOAT NOT NULL,
    `date` DATE NOT NULL,
    accountID INT,

    PRIMARY KEY (id),
    FOREIGN KEY (accountID) REFERENCES accounts(id)
);