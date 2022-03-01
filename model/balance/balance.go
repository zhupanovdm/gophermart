package balance

import "github.com/zhupanovdm/gophermart/model"

type Balance struct {
	Current   model.Sum `json:"current"`
	Withdrawn model.Sum `json:"withdrawn"`
}
