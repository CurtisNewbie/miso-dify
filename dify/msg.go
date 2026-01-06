package dify

import (
	"fmt"

	"github.com/curtisnewbie/miso/errs"
	"github.com/curtisnewbie/miso/miso"
)

const (
	RatingLike    = "like"
	RatingDislike = "dislike"
)

type MsgFeedbackReq struct {
	MessageId string `json:"-"`
	Rating    string `json:"rating"`
	User      string `json:"user"`
	Content   string `json:"content"`
}

type apiMsgFeedbackReq struct {
	Rating  *string `json:"rating"`
	User    string  `json:"user"`
	Content string  `json:"content"`
}

func SendMsgFeedback(rail miso.Rail, host string, apiKey string, req MsgFeedbackReq) error {
	url := host + fmt.Sprintf("/v1/messages/%v/feedbacks", req.MessageId)
	var rating *string = nil // nil: cancel rating
	if req.Rating != "" {
		rating = &req.Rating
	}
	s, err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(apiMsgFeedbackReq{
			User:    req.User,
			Rating:  rating,
			Content: req.Content,
		}).
		Str()
	if err != nil {
		return errs.Wrapf(err, "dify SendMsgFeedback failed")
	}
	rail.Infof("Request success, %v", s)
	return nil
}
