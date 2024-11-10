package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type bank_preset struct {
	csv_format  map[string]int
	date_format string
}

var ( // bank presets
	starling_csv  = map[string]int{"Date": 0, "Party": 1, "Reference": 2, "PaidIn": 4, "PaidOut": -1, "Balance": 5}
	starling_date = "02/01/2006"

	hsbc_csv  = map[string]int{"Date": 0, "Party": -1, "Reference": 2, "PaidIn": 2, "PaidOut": 3, "Balance": -1}
	hsbc_date = "02 Jan 06"

	//bank map
	banks = map[string]bank_preset{
		"starling": {csv_format: starling_csv, date_format: starling_date},
		"hsbc":     {csv_format: hsbc_csv, date_format: hsbc_date},
	}
)

var (
	csv_format      map[string]int //= map[string]int{"Date": 0, "Party": 1, "Reference": -1, "PaidIn": 4, "PaidOut": -1, "Balance": 5}
	date_format     string         // = "02/01/2006"
	input_path      = flag.String("p", "", "CSV Path")
	input_bank      = flag.String("b", "", "Input bank")
	input_user      = flag.Int("u", -1, "User ID")
	input_operation = flag.Bool("o", false, " If false Parse CSV; require CSV path, bank, userID, If true send prompt require userID")
	input_testmode  = flag.Bool("t", false, "Enter test mode to run client")
	open_ai_key     string
)

type transaction struct {
	date      time.Time // when
	party     string    //who to
	reference string    // what for
	amount    float64   //how much
	balance   float64
}

type statement struct {
	transactions []transaction
	start        time.Time
	period       time.Duration
}

func main() {

	flag.Parse()

	if *input_testmode {

	}

	switch {
	case *input_operation: // send prompt o is true

		err := godotenv.Load() // load api key
		if err != nil {
			fmt.Printf("Error opening .env file %v\n", err)
		}

		open_ai_key = os.Getenv("OPENAI_API_KEY")

		if open_ai_key == "" {
			log.Fatalf("API key not found")
		}
		if !*input_testmode {
			summarise(read_from_db(*input_user), open_ai_key)
		} else {
			//summarise(read_from_db(*input_user), open_ai_key) unsure

		}

	case !*input_operation: //parse csv o is false

		preset, ok := banks[*input_bank]
		if !ok {
			log.Fatalf("Bank not valid")
		}

		csv_format, date_format = preset.csv_format, preset.date_format

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
		if !*input_testmode {
			stmnt := parse_csv(records)
			plot := Create_plot(&stmnt)
			Plot_to_csv(plot)
			save_to_db(stmnt, *input_user) // parse -> save to db
		} else {

			//TEST MODE
			fmt.Print("Running Test Mode\n")
			err := godotenv.Load() // load api key
			if err != nil {
				fmt.Printf("Error opening .env file %v\n", err)
			}

			open_ai_key = os.Getenv("OPENAI_API_KEY")
			if open_ai_key == "" {
				log.Fatalf("API key not found")
			}
			summarise(parse_csv(records).transactions, open_ai_key)

		}
	}

	//summarise(parse_csv(records).transactions, open_ai_key)
}

func parse_csv(records [][]string) statement {

	var transactions []transaction
	var end_date time.Time
	var start_date time.Time

	verify_index_float := func(name string, record []string) (float64, error) {
		var err error
		var value float64
		index := csv_format[name]

		if index != -1 && index < len(record) {
			str_val := record[index]
			if str_val == "" {
				return 0, nil // Return 0 for missing or empty values
			}
			value, err = strconv.ParseFloat(str_val, 64)
		}
		return value, err
	}

	verify_index_string := func(name string, record []string) string {
		var index int
		if index = csv_format[name]; index == -1 {
			return ""
		}
		return record[index]
	}

	//first row is header

	for i, record := range records {
		//fmt.Printf(record[0] + "	" + record[1] + "	" + record[2] + "	" + record[3] + "	" + record[4] + "\n")
		if i == 0 {
			continue
		}

		// Parse the data
		date_index := csv_format["Date"]
		date, err := time.Parse(date_format, record[date_index]) // date
		if err != nil {
			fmt.Printf("Date: Error parsing date on row %d: %v\n", i+1, err)
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

		// account is a string

		// company is a string

		// Place is a string

		// Reference is a string

		out, err := verify_index_float("PaidOut", record)
		if err != nil {
			fmt.Printf("Out: Error parsing amount on row %d: %v\n", i+1, err)
			continue
		}

		in, err := verify_index_float("PaidIn", record)
		if err != nil {
			fmt.Printf("In: Error parsing amount on row %d: %v\n", i+1, err)
			continue
		}

		balance, err := verify_index_float("Balance", record)
		if err != nil {
			fmt.Printf("Bal: Error parsing amount on row %d: %v\n", i+1, err)
			continue
		}

		// build transaction
		trans := transaction{
			date:      date,
			party:     verify_index_string("Party", record),
			reference: verify_index_string("Reference", record),
			amount:    in - out,
			balance:   balance,
		}
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

	fmt.Printf("Statement Start Date: %v\n", stmt.start)
	fmt.Printf("Statement Period: %v days\n", stmt.period.Hours()/24)
	fmt.Printf("Transactions: %d \n", len(stmt.transactions))
	//for _, trans := range transactions {
	//	fmt.Println(trans)
	//}

	return stmt
}

func save_to_db(stmt statement, userid int) {
	connect_start := "user=admin dbname=app password=password host=127.0.0.1 port=6543"
	db, err := sql.Open("postgres", connect_start)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	for _, trans := range stmt.transactions {
		date := trans.date.Format("2006-01-02")
		party, reference, amount, balance := trans.party, trans.reference, trans.amount, trans.balance
		_, err := db.Exec(`INSERT INTO Transactions (userid, date, party, reference, amount, balance)
							VALUES ($1, $2, $3, $4 ,$5, $6)`,
			userid, date, party, reference, amount, balance)
		if err != nil {
			log.Printf("error inserting transaction on date %s", date)
		}
	}
	fmt.Printf("Saved to the database sucessfully")

}

func get_transactions(db *sql.DB, userid int) ([]transaction, error) {
	query := `SELECT * FROM  transactions WHERE userid = $1`
	row, err := db.Query(query, userid)
	if err != nil {
		fmt.Printf("error querrying transactions %v", err)
		return nil, err
	}
	defer row.Close()

	var transactions []transaction

	for row.Next() {
		var trans transaction
		err := row.Scan(&trans.date, &trans.party, &trans.reference, &trans.amount, &trans.balance)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		transactions = append(transactions, trans)
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration %w", err)
	}

	return transactions, nil
}

func read_from_db(userid int) []transaction {
	connect_start := "user=admin dbname=app password=password host=127.0.0.1 port=6543"
	db, err := sql.Open("postgres", connect_start)
	if err != nil {
		log.Fatal("error connecting to database", err)
	}
	defer db.Close()

	transactions, err := get_transactions(db, userid)
	if err != nil {
		log.Fatal("error fetching transactions:", err)
	}
	return transactions
}

func summarise(statement []transaction, key string) {

	prompt := buildPrompt(statement)

	client := openai.NewClient(key)

	// Create a context to use with the API call
	ctx := context.Background()

	// Make the API call using the context
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		fmt.Printf("Error making OpenAI API call: %v\n", err)
		return
	}

	// Print the response
	fmt.Println("Response from OpenAI:")
	fmt.Println(resp.Choices[0].Message.Content)
}

func buildPrompt(transactions []transaction) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("Analyze the following bank statement data:\n\n")
	for _, txn := range transactions {
		promptBuilder.WriteString(fmt.Sprintf(
			"Date: %s, Party: %s, Reference: %s, Amount: %.2f, Balance: %.2f\n",
			txn.date.Format("2006-01-02"), txn.party, txn.reference, txn.amount, txn.balance,
		))
	}
	promptBuilder.WriteString("\nProvide the following insights:\n")
	promptBuilder.WriteString("- Rigid and flexible spending habits\n")
	promptBuilder.WriteString("- Highlights of potential saving opportunities\n")
	promptBuilder.WriteString("- Calculate a 'budgeting score' which takes in values from income, saving goal, the amount and frequency of flexible expenses. This goal should be out of 100, where 0 is 100% impulsive. Impulsivity should lower the score.")

	return promptBuilder.String()
}
