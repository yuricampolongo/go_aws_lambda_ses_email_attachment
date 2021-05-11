# GO AWS Lambda email attachment

Send email with attachment using an AWS Lambda Function and AWS Simple Email Service

## Get started

All the code is in the main.go file, I've done that because it is a simple function and to keep easier to compile and send to AWS Lambda Function, but you can break in smaller pieces if you want 

## Configuration

All the configuration can be found in the const area of the main.go file.

 - AWS_REGION = The region that your AWS account is in (default 'us-east-2')
 - FROM = The email address to be used as a sender
 - AWS_ACCESS_KEY = aws IAM access key from a user that has access to send emails on the AWS SES resource
 - AWS_SECRET_KEY = aws IAM secret key from a user that has access to send emails on the AWS SES resource
 - PDF_TEMPLATE_FILE = Image to be used as a template to the pdf that will be send as an attachment.
 - SUBJECT = Email subject
 - BODY = Email body

You can delete the AWS_ACCESS_KEY and AWS_SECRET_KEY if you have your credentials stored in your enviroment variables or in your ~/aws/credentials file. I put in the variables because in my case, this code will be runing inside an serverless environment.

 - Event Struct - You can modify this struct to insert more fields that your aws lambda function will pass as JSON, in my case I'll receive only the e-mail, so only one field is enough for me.

## Compilation

Navigate to the folder that contains your main.go file and run the following commands

```go build main.go```

```zip -r function.zip .```

Upload your .zip file to your AWS Lambda

## Test AWS Lambda Function

Create a test event in your AWS Lambda function, the body must match your Event struct as a JSON string:

Example: 
```
{
  "email": "yuricampolongo@outlook.com"
}
```