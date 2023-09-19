package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

type S3ImageDetailsRequest struct {
	Bucket        string `json:"bucket"`
	SelfieImage   string `json:"selfieImage"`
	DocumentImage string `json:"documentImage"`
}

type Label struct {
	Confidence float64
	Name       string
}

type Result struct {
	SelfieDetails         rekognition.FaceDetail         `json:"selfieDetails"`
	DocumentFaceDetails   rekognition.FaceDetail         `json:"documentFaceDetails"`
	SelfieMatchesDocument rekognition.CompareFacesOutput `json:"selfieMatchesDocument"`
}

type Response struct {
	Message string `json:"message"`
	Result  Result `json:"result"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var s3ImageDetails S3ImageDetailsRequest

	err := json.Unmarshal([]byte(request.Body), &s3ImageDetails)

	if err != nil {
		log.Println("Error unmarshalling request body:", err)

		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request body",
		}, nil
	}

	session, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		log.Println("Failed to create AWS session:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error creating rekognition session : %v", err)
	}

	svc := rekognition.New(session)

	checkSelfieInput := &rekognition.DetectFacesInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(s3ImageDetails.Bucket),
				Name:   aws.String(s3ImageDetails.SelfieImage),
			},
		},
	}

	checkSelfie, err := svc.DetectFaces(checkSelfieInput)
	if err != nil {
		log.Println("Failed to detect labels:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error check selfie image: %v", err)
	}

	var selfieDetails rekognition.FaceDetail

	if len(checkSelfie.FaceDetails) > 0 {
		// Assume there is one face in the image
		selfieDetails = *checkSelfie.FaceDetails[0]

	}

	checkIdentityDocumentInput := &rekognition.DetectFacesInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(s3ImageDetails.Bucket),
				Name:   aws.String(s3ImageDetails.DocumentImage),
			},
		},
	}

	checkIdentityDocument, err := svc.DetectFaces(checkIdentityDocumentInput)

	if err != nil {
		log.Println("Failed to detect labels:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error failed to check identity document : %v", err)

	}

	var identityDocumentFaceDetails rekognition.FaceDetail

	if len(checkIdentityDocument.FaceDetails) > 0 {
		// Assume there is one face in the image
		identityDocumentFaceDetails = *checkIdentityDocument.FaceDetails[0]

	}

	compareSelfieWithIDDocumentInput := &rekognition.CompareFacesInput{
		SourceImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(s3ImageDetails.Bucket),
				Name:   aws.String(s3ImageDetails.DocumentImage),
			},
		},
		TargetImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(s3ImageDetails.Bucket),
				Name:   aws.String(s3ImageDetails.SelfieImage),
			},
		},
	}

	compareSelfieWithIdDocument, err := svc.CompareFaces(compareSelfieWithIDDocumentInput)

	if err != nil {
		log.Println("Failed to detect labels:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error failed to compare selfie with identity document : %v", err)
	}

	response := Response{
		Message: "KYC Documents Analysis Results",
		Result: Result{
			SelfieDetails:         selfieDetails,
			DocumentFaceDetails:   identityDocumentFaceDetails,
			SelfieMatchesDocument: *compareSelfieWithIdDocument,
		},
	}

	body, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error marshalling response: %v", err)
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
