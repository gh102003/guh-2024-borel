#!/bin/bash
psql postgresql://user:password@127.0.0.1:6543 << EOF
    CREATE DATABASE app
EOF
psql postgresql://user:password@127.0.0.1:6543/app << EOF
    CREATE TABLE Transactions (
    serial_id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    account VARCHAR(50) NOT NULL,
    company VARCHAR(100),
    location VARCHAR(100),
    reference VARCHAR(100),
    amount DECIMAL(15, 2) NOT NULL,
    balance DECIMAL(15, 2) NOT NULL,
	  userid INT
    );
    CREATE TABLE Users (
    userid INT PRIMARY KEY DEFAULT floor(random() * 1000000)
    );
EOF
