package apigatewayutils

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

func Redirect(to string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": to,
		},
		Body: fmt.Sprintf("Redirecting to %s", to),
	}
}

func NotFound(msg string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       msg,
	}
}

func NoContent() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}
}

func Created() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
	}
}

func BadRequest(msg string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       msg,
	}
}

func InternalServerError(msg string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       msg,
	}
}

func RespondJson(out interface{}) (*events.APIGatewayProxyResponse, error) {
	asJson, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(asJson),
	}, nil
}

func OkText(msg string) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: string(msg),
	}, nil
}
