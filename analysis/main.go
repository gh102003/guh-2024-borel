package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type bank_preset struct {
	csv_format  map[string]int
	date_format string
}

var ( // bank presets
	starling_csv  = map[string]int{"Date": 0, "Account": -1, "Company": 1, "Location": -1, "PaidIn": 4, "PaidOut": -1, "Balance": 5}
	starling_date = "02/01/2006"

	hsbc_csv  = map[string]int{"Date": 0, "Account": 1, "Company": -1, "Location": -1, "PaidIn": 3, "PaidOut": 4, "Balance": -1}
	hsbc_date = "02 Jan 06"

	//bank map
	banks = map[string]bank_preset{
		"starling": {csv_format: starling_csv, date_format: starling_date},
		"hsbc":     {csv_format: hsbc_csv, date_format: hsbc_date},
	}
)

var (
	csv_format  = map[string]int{"Date": 0, "Account": -1, "Company": 1, "Location": -1, "PaidIn": 4, "PaidOut": -1, "Balance": 5}
	date_format = "02/01/2006"
	input_path  = flag.String("p", "", "CSV Path")
	input_bank  = flag.String("b", "", "Input bank")
	open_ai_key string
)

type transaction struct {
	date      time.Time
	account   string
	company   string
	location  string
	reference string
	in        float64
	out       float64
	balance   float64
}

type statement struct {
	transactions []transaction
	start        time.Time
	period       time.Duration
}

func main() {

	flag.Parse()

	preset := banks[*input_bank]
	csv_format, date_format = preset.csv_format, preset.date_format //handle exceptions later

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error opening .env file %v\n", err)
	}

	open_ai_key = os.Getenv("OPENAI_API_KEY")
	if open_ai_key == "" {
		log.Fatalf("API key not found")
	}

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

	summarise(parse_csv(records), open_ai_key)
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
			account:   verify_index_string("Account", record),
			company:   verify_index_string("Company", record),
			location:  verify_index_string("Location", record),
			reference: verify_index_string("Reference", record),
			out:       out,
			in:        in,
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

	return stmt
}

func summarise(statement statement, key string) {

	prompt := buildPrompt(statement.transactions)

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
			"Date: %s, Account: %s, Company: %s, Location: %s, Reference: %s, In: %.2f, Out:%.2f, Balance: %.2f\n",
			txn.date.Format("2006-01-02"), txn.account, txn.company, txn.location, txn.reference, txn.in, txn.out, txn.balance,
		))
	}
	promptBuilder.WriteString("\nProvide the following insights:\n")
	promptBuilder.WriteString("- Account summary (total balance, total inflow, total outflow)\n")
	promptBuilder.WriteString("- Rigid and flexible spending habits\n")
	promptBuilder.WriteString("- Insights into top spending categories and most frequent transactions\n")
	promptBuilder.WriteString("- Highlights of potential saving opportunities\n")

	return promptBuilder.String()
}

/*
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
*/
/* func catagorise(data statement, catagories map[string]string) [][]transaction {

	for transaction := range data.transactions {

	}

}
*/
