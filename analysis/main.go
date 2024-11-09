package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	input_path = flag.String("p", "", "CSV Path")
)

type transaction struct {
	date      time.Time
	account   string
	company   string
	location  string
	reference string
	amount    float64
	balance   float64
}

type statement struct {
	transactions []transaction
	start        time.Time
	period       time.Duration
}

func main() {

	flag.Parse()

	file, err := os.Open(*input_path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}

	var transactions []transaction
	var end_date time.Time
	var start_date time.Time

	//first row is header
	for i, record := range records {
		if i == 0 {
			continue
		}

		// Parse the data
		date, err := time.Parse("2006-01-02", record[0]) // date
		if err != nil {
			fmt.Printf("Error parsing date on row %d: %v\n", i+1, err)
			continue
		}

		// account is a string

		// company is a string

		// Place is a string

		// Reference is a string

		amount, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			fmt.Printf("Error parsing amount on row %d: %v\n", i+1, err)
			continue
		}

		balance, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			fmt.Printf("Error parsing amount on row %d: %v\n", i+1, err)
			continue
		}

		if i == 1 {
			start_date = date
			end_date = date
		}

		if date.Before(start_date) {
			start_date = date
		}
		if date.After(end_date) {
			end_date = date
		}

		trans := transaction{
			date:      date,
			account:   record[1],
			company:   record[2],
			location:  record[3],
			reference: record[4],
			amount:    amount,
			balance:   balance,
		}

		// Add the transaction to the slice
		transactions = append(transactions, trans)
	}

	// Calculate the period duration
	period := end_date.Sub(start_date)

	// Create a statement
	stmt := statement{
		transactions: transactions,
		start:        start_date,
		period:       period,
	}

	// Print the parsed statement (for debugging)
	fmt.Printf("Statement Start Date: %v\n", stmt.start)
	fmt.Printf("Statement Period: %v days\n", stmt.period.Hours()/24)
	fmt.Printf("Transactions: %d \n", len(stmt.transactions))
	/* for _, txn := range stmt.transactions {
	fmt.Printf("Date: %v, Company: %s, Location: %s, Reference: %s, Amount: %.2f\n",
		txn.date, txn.company, txn.location, txn.reference, txn.amount) */
}

func (data *statement) total_spend() float64 {
	var sum float64 = 0
	for _, trans := range data.transactions {
		sum += trans.amount
	}
	return sum
}

func (data *statement) total_in() float64 {
	var sum float64 = 0
	for _, trans := range data.transactions {
		if trans.amount > 0 {
			sum += trans.amount
		}
	}
	return sum
}

func (data *statement) total_out() float64 {
	return data.total_spend() - data.total_in()
}

func (data *statement) total_trans() int {
	return len(data.transactions)
}

func (data *statement) frequancy() float64 {
	trans := data.total_trans()
	return float64(trans) / data.period.Hours() / float64(24)

}

func (data *statement) summarise() {
	fmt.Printf("Summary for statement from %f to %f", float64(data.start.Hour())/float64(24), float64(data.start.Add(data.period).Hour())/float64(24))
	fmt.Printf("Total Transactions: %d \n Transaction Frequancy: %f \n Income: %f \n Expenses: %f", data.total_trans(), data.frequancy(), data.total_in(), data.total_out())
}

/* func catagorise(data statement, catagories map[string]string) [][]transaction {

	for transaction := range data.transactions {

	}

}
*/
