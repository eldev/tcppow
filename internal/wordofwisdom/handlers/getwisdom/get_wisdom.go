package getwisdom

import (
	"encoding/json"
	"log"
	"tcppow/internal/netprotocol"
	"tcppow/internal/wordofwisdom/domain"

	"github.com/pkg/errors"
)

type Handler struct {
	businessLogic *domain.Model
}

func New(businessLogic *domain.Model) *Handler {
	return &Handler{
		businessLogic: businessLogic,
	}
}

type Response struct {
	WordOfWisdom string `json:"word_of_wisdom"`
}

func (h *Handler) HandlerFunc() func(resp *netprotocol.Response, req *netprotocol.Request) error {
	return func(resp *netprotocol.Response, req *netprotocol.Request) error {

		log.Printf("get-wisdom\n")

		wisdomWord, err := h.businessLogic.GetWordOfWisdom(req.RequestContext)
		if err != nil {
			resp.Status = netprotocol.INTERNAL_SERVER_ERROR
			return errors.WithMessage(err, "getting word of wisdom")
		}

		response := Response{
			WordOfWisdom: wisdomWord,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			resp.Status = netprotocol.INTERNAL_SERVER_ERROR
			return errors.WithMessage(err, "json marshalling response")
		}

		resp.Status = 200
		resp.Body = jsonResponse

		return nil
	}
}
