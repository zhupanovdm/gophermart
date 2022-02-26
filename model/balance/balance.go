package balance

import "github.com/zhupanovdm/gophermart/model"

type Balance struct {
	Current   model.Money `json:"current"`
	Withdrawn model.Money `json:"withdrawn"`
}
