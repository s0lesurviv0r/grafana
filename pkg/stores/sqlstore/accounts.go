package sqlstore

import (
	"time"

	"github.com/go-xorm/xorm"

	"github.com/torkelo/grafana-pro/pkg/bus"
	m "github.com/torkelo/grafana-pro/pkg/models"
)

func init() {
	bus.AddHandler("sql", GetAccountInfo)
	bus.AddHandler("sql", AddCollaborator)
}

func GetAccountInfo(query *m.GetAccountInfoQuery) error {
	var account m.Account
	has, err := x.Id(query.Id).Get(&account)

	if err != nil {
		return err
	} else if has == false {
		return m.ErrAccountNotFound
	}

	query.Result = m.AccountDTO{
		Name:          account.Name,
		Email:         account.Email,
		Collaborators: make([]*m.CollaboratorDTO, 0),
	}

	sess := x.Table("collaborator")
	sess.Join("INNER", "account", "account.id=collaborator.account_Id")
	sess.Where("collaborator.for_account_id=?", query.Id)
	err = sess.Find(&query.Result.Collaborators)

	return err
}

func AddCollaborator(cmd *m.AddCollaboratorCommand) error {
	return inTransaction(func(sess *xorm.Session) error {

		entity := m.Collaborator{
			AccountId:    cmd.AccountId,
			ForAccountId: cmd.ForAccountId,
			Role:         cmd.Role,
			Created:      time.Now(),
			Updated:      time.Now(),
		}

		_, err := sess.Insert(&entity)
		return err
	})
}

func SaveAccount(account *m.Account) error {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	if account.Id == 0 {
		_, err = sess.Insert(account)
	} else {
		_, err = sess.Id(account.Id).Update(account)
	}

	if err != nil {
		sess.Rollback()
		return err
	} else if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}

func GetAccount(id int64) (*m.Account, error) {
	var err error

	var account m.Account
	has, err := x.Id(id).Get(&account)

	if err != nil {
		return nil, err
	} else if has == false {
		return nil, m.ErrAccountNotFound
	}

	if account.UsingAccountId == 0 {
		account.UsingAccountId = account.Id
	}

	return &account, nil
}

func GetAccountByLogin(emailOrLogin string) (*m.Account, error) {
	var err error

	account := &m.Account{Login: emailOrLogin}
	has, err := x.Get(account)

	if err != nil {
		return nil, err
	} else if has == false {
		return nil, m.ErrAccountNotFound
	}

	return account, nil
}

func GetOtherAccountsFor(accountId int64) ([]*m.OtherAccount, error) {
	collaborators := make([]*m.OtherAccount, 0)
	sess := x.Table("collaborator")
	sess.Join("INNER", "account", "collaborator.for_account_id=account.id")
	sess.Where("account_id=?", accountId)
	sess.Cols("collaborator.id", "collaborator.role", "account.email")
	err := sess.Find(&collaborators)
	return collaborators, err
}
