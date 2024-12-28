# AWS Lambda Deployment Guide

This README provides a step-by-step guide to deploy a Go-based AWS Lambda function. The function accepts a JSON payload with two fields, `account_id` and `connection_id`. If these fields are missing, it returns an unauthorized error. Otherwise, it retrieves two environment variables (`KEY_ID` and `KEY_SECRET`) and returns them in the response.

---

## Steps to Deploy

### 1. Build the Go Lambda Function
Compile the Go Lambda function into a Linux-compatible binary:

```bash
GOOS=linux GOARCH=arm64 go build -o bootstrap main.go 
```

This will create an executable named `bootstrap`.

---

### 2. Package the Function

Create a `.zip` file containing the binary:

```bash
zip function.zip bootstrap
```

---

### 3. Deploy the Lambda Function

Deploy the function to AWS:

```bash
aws lambda create-function \
  --function-name go-lambda-example \
  --runtime go1.x \
  --handler main \
  --role arn:aws:iam::<ACCOUNT_ID>:role/LambdaExecutionRole \
  --zip-file fileb://function.zip
```

Replace `<ACCOUNT_ID>` with your AWS account ID.

---

### 5. Set Environment Variables

Set the `USER` and `PASS` environment variables:

```bash
aws lambda update-function-configuration \
  --function-name go-lambda-example \
  --environment "Variables={KEY_ID=admin,KEY_SECRET=secret}"
```