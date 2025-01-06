# AWS Lambda Deployment Guide

The following Lambda function is responsible for fetching billing-related data in environments where the customer is unable to grant billing-related permissions to the IAM user or role assigned to the Superstream platform.

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

Upload the Zip file from the AWS Lambda console

---

### 4. Create an IAM role and attach it to the Lambda function

The role should have the following permissions

```bash
"kafka:ListTagsForResource"
"kafka:TagResource"
"ce:ListCostAllocationTags"
"ce:UpdateCostAllocationTagsStatus"
"ce:GetCostAndUsageWithResources"
"ce:GetCostAndUsage"
"ce:ListTagsForResource"
"ce:GetTags"
 +
AWSBillingReadOnlyAccess policy
```
