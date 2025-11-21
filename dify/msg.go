package dify

import (
	"fmt"

	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util/errs"
)

const (
	RatingLike    = "like"
	RatingDislike = "dislike"
)

type MsgFeedbackReq struct {
	MessageId string `json:"-"`
	Rating    string
	User      string
	Content   string
}

type apiMsgFeedbackReq struct {
	Rating  *string
	User    string
	Content string
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
