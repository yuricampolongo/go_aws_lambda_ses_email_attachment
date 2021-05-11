package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/signintech/gopdf"
)

const (
	AWS_REGION        = "us-east-2"
	FROM              = "your_email@here"
	AWS_ACCESS_KEY    = ""
	AWS_SECRET_KEY    = ""
	PDF_TEMPLATE_FILE = "template.jpg"
	SUBJECT           = "Your subject here"
	BODY              = "Your email body here"
)

var (
	svc *ses.SES
)

// This struct must be compatible with the json received from your AWS Lambda function
type Event struct {
	Email string `json:"email"`
}

/**
 * Initializes the AWS session to send the email using AWS SES
 * Your AWS user that generated the ACCESS_KEY and SECRET_KEY
 * must have the permissions to send email on AWS SES resource
 */
func init() {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY, AWS_SECRET_KEY, ""),
	})
	if err != nil {
		fmt.Println("unable to connect to email provider")
	}

	svc = ses.New(sess)
}

func main() {
	lambda.Start(HandleRequest)
}

/**
 * Handles the lambda function, the input must be a valid json from the struct
 * Event, declared above. If you need more information from your LambdaFunction
 * add more fields on the struct Event
 */
func HandleRequest(ctx context.Context, invite Event) (string, error) {
	pdf := PreparePdf()
	SendEmail(SUBJECT, BODY, invite.Email, pdf)
	return "email sent to " + invite.Email, nil
}

/**
 * Prepares the PDF to send as an attachment
 *
 * The PDF generate contains only an image as a template that occupies the entire A4 Page
 */
func PreparePdf() *gopdf.GoPdf {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	pdf.Image(PDF_TEMPLATE_FILE, 0, 0, &gopdf.Rect{W: 650, H: 900})
	return &pdf
}

/**
 * Sends the email to the addres received as a parameter from the aws lambda function
 */
func SendEmail(subject string, body string, to string, pdf *gopdf.GoPdf) (bool, error) {
	file := pdf.GetBytesPdf()
	input, _ := buildEmailInput(FROM, to, subject, body, file)
	_, err := svc.SendRawEmail(input)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

/**
 * Build the AWS Simple Email Service input with all the email parts
 */
func buildEmailInput(source, destination, subject, message string, file []byte) (*ses.SendRawEmailInput, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// email main header:
	h := make(textproto.MIMEHeader)
	h.Set("From", source)
	h.Set("To", destination)
	h.Set("Return-Path", source)
	h.Set("Subject", subject)
	h.Set("Content-Language", "en-US")
	h.Set("Content-Type", "multipart/mixed; boundary=\""+writer.Boundary()+"\"")
	h.Set("MIME-Version", "1.0")
	_, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	// body:
	h = make(textproto.MIMEHeader)
	h.Set("Content-Transfer-Encoding", "7bit")
	h.Set("Content-Type", "text/plain; charset=us-ascii")
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}
	_, err = part.Write([]byte(message))
	if err != nil {
		return nil, err
	}

	// file attachment:
	fn := "invite.pdf"
	h = make(textproto.MIMEHeader)
	h.Set("Content-Type", "application/pdf; charset=us-ascii")
	h.Set("Content-Transfer-Encoding", "base64")
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fn))
	fileEncoded := base64.StdEncoding.EncodeToString(file)
	part, err = writer.CreatePart(h)
	if err != nil {
		return nil, err
	}
	_, err = part.Write([]byte(fileEncoded))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Strip boundary line before header (doesn't work with it present)
	s := buf.String()
	if strings.Count(s, "\n") < 2 {
		return nil, fmt.Errorf("invalid e-mail content")
	}
	s = strings.SplitN(s, "\n", 2)[1]

	raw := ses.RawMessage{
		Data: []byte(s),
	}
	input := &ses.SendRawEmailInput{
		Destinations: []*string{aws.String(destination)},
		Source:       aws.String(source),
		RawMessage:   &raw,
	}

	return input, nil
}
