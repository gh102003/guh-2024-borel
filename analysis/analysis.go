package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

func Quicksort(transactions []transaction) { // we require a sorted array for volatility analysis

	if len(transactions) <= 1 { //implement quick sort
		return
	}

	pivot := transactions[len(transactions)-1]
	left := 0
	right := len(transactions) - 2

	for left <= right {

		for left <= right && transactions[left].date.Before(pivot.date) {
			left++
		}
		for left <= right && transactions[right].date.After(pivot.date) {
			right--
		}
		if left <= right {
			transactions[left], transactions[right] = transactions[right], transactions[left]
			left++
			right--
		}
	}

	transactions[left], transactions[len(transactions)-1] = transactions[len(transactions)-1], transactions[left]

	Quicksort(transactions[:left])
	Quicksort(transactions[left+1:])

}

func Trans_to_statement(transactions []transaction) statement {
	Quicksort(transactions)
	end := transactions[len(transactions)-1]
	start := transactions[0]

	stmnt := statement{transactions: transactions, start: start.date, period: end.date.Sub(start.date)}

	return stmnt
}

func Create_plot(statement *statement) map[int]float64 { // maybe george can do something with this
	current := statement.start
	pos := 0
	plot := make(map[int]float64)

	for _, trans := range statement.transactions {

		if trans.date.After(current) {
			pos += 10 * int((trans.date.Sub(current).Hours())/float64(24)) // the day difference between the points
			plot[pos] = trans.amount
		} else {
			pos += 1
			plot[pos] = trans.amount
		}

	}
	return plot

}

// "small" transactions are oens under the threashhold - for example <$10
// find standard deviation of small transactions - called the volatility metric
func Volatility(stmnt statement, threshhold int) float64 {
	var small []float64

	for _, trans := range stmnt.transactions { // threashhold cutoff
		if trans.amount < float64(threshhold) {
			small = append(small, trans.amount)
		}
	}

	var sum float64 = 0
	var variance float64 = 0

	for _, x := range small {
		sum += x
	}
	mean := sum / float64(len(small))

	for _, x := range small {
		variance = (x - mean) * (x - mean)
	}
	variance = variance / float64(len(small))

	return math.Sqrt(variance)
}

// Impulsivity is measured by high frequancy of small transactions
func Impulsivity(stmnt statement, threshhold int) float64 {
	return 0
}

func Plot_to_csv(plot map[int]float64) {

	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"Key", "Value"})
	if err != nil {
		fmt.Printf("Error writing header: %v\n", err)
		return
	}

	for key, value := range plot {
		record := []string{strconv.Itoa(key), strconv.FormatFloat(value, 'f', 2, 64)}
		err := writer.Write(record)
		if err != nil {
			fmt.Printf("Error writing record: %v\n", err)
			return
		}
	}
}
