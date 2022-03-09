package trxcontext

import (
	"example.com/TokenCC/lib/account"
	"example.com/TokenCC/lib/admin"
	"example.com/TokenCC/lib/auth"
	"example.com/TokenCC/lib/holding"
	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/role"
	"example.com/TokenCC/lib/token"
	"example.com/TokenCC/lib/transaction"
	"example.com/TokenCC/lib/signature"
	"example.com/TokenCC/lib/user"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type TrxContext struct {
	Model       *model.Model
	Token       *token.TokenReciever
	Role        *role.RoleReciever
	Auth        *auth.AuthReciever
	Hold        *holding.HoldReciever
	Admin       *admin.AdminReciever
	Account     *account.AccountReciever
	Transaction *transaction.TransactionReciever
	TxSignature *signature.SignatureReciever
	User        *user.UserReciever
}

func GetNewCtx(stub shim.ChaincodeStubInterface) TrxContext {
	var ctx TrxContext
	ctx.Model = model.GetNewModel(stub)
	ctx.Transaction = transaction.GetNewTransactionReciever(ctx.Model)
	ctx.Account = account.GetNewAccountReciever(ctx.Model, ctx.Transaction)
	ctx.Hold = holding.GetNewHoldReciever(ctx.Model)
	ctx.Admin = admin.GetNewAdminReciever(ctx.Model)
	ctx.Role = role.GetNewRoleReciever(ctx.Model, ctx.Account)
	ctx.Auth = auth.GetNewAuthReciever(ctx.Model, ctx.Admin, ctx.Account)
	ctx.Token = token.GetNewTokenReciever(ctx.Model, ctx.Role, ctx.Hold, ctx.Account, ctx.Transaction)
	ctx.Transaction = transaction.GetNewTransactionReciever(ctx.Model)
	ctx.TxSignature = signature.GetNewSignatureReciever(ctx.Model, ctx.Transaction, ctx.Account)
	ctx.User = user.GetNewUserReciever(ctx.Model)

	return ctx
}
