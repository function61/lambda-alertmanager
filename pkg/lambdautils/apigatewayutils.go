package lambdautils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apex/gateway"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

// github.com/akrylysov/algnhsa has similar implementation than apex/gateway, but had the
// useful bits non-exported and it used httptest for production code
func ServeApiGatewayProxyRequestUsingHttpHandler(
	ctx context.Context,
	proxyRequest *events.APIGatewayProxyRequest,
	httpHandler http.Handler,
) ([]byte, error) {
	request, err := gateway.NewRequest(ctx, *proxyRequest)
	if err != nil {
		return nil, err
	}

	response := gateway.NewResponse()

	httpHandler.ServeHTTP(response, request)

	proxyResponse := response.End()

	return json.Marshal(&proxyResponse)
}

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
