package view

import (
	"errors"
	"fmt"

	"github.com/crypto-com/chain-indexing/usecase/coin"

	"github.com/crypto-com/chain-indexing/appinterface/projection/view"

	"github.com/crypto-com/chain-indexing/appinterface/pagination"
	"github.com/crypto-com/chain-indexing/appinterface/rdb"
	_ "github.com/crypto-com/chain-indexing/test/factory"
)

type Accounts struct {
	rdb *rdb.Handle
}

type AccountsListOrder struct {
	AccountAddress view.ORDER
}

type AccountIdentity struct {
	MaybeAddress string
}

func NewAccounts(handle *rdb.Handle) *Accounts {
	return &Accounts{
		handle,
	}
}

func (accountsView *Accounts) Upsert(account *AccountRow) error {
	var err error
	var sql string
	sql, _, err = accountsView.rdb.StmtBuilder.
		Insert(
			"view_accounts",
		).
		Columns(
			"account_type",
			"account_address",
			"pubkey",
			"account_number",
			"sequence_number",
			"balance",
		).
		Values("?", "?", "?", "?", "?", "?").
		Suffix("ON CONFLICT(account_address) DO UPDATE SET balance = EXCLUDED.balance").
		ToSql()

	if err != nil {
		return fmt.Errorf("error building accounts insertion sql: %v: %w", err, rdb.ErrBuildSQLStmt)
	}

	result, err := accountsView.rdb.Exec(sql,
		account.Type,
		account.Address,
		account.Pubkey,
		account.AccountNumber,
		account.SequenceNumber,
		account.Balance.String(),
	)
	if err != nil {
		return fmt.Errorf("error inserting block into the table: %v: %w", err, rdb.ErrWrite)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("error inserting block into the table: no rows inserted: %w", rdb.ErrWrite)
	}

	return nil
}

func (accountsView *Accounts) FindBy(identity *AccountIdentity) (*AccountRow, error) {
	var err error

	selectStmtBuilder := accountsView.rdb.StmtBuilder.Select(
		"account_type",
		"account_address",
		"pubkey",
		"account_number",
		"sequence_number",

		"account_balance", "account_denom",
	).From("view_accounts")

	selectStmtBuilder = selectStmtBuilder.Where("account_address = ?", identity.MaybeAddress)

	sql, sqlArgs, err := selectStmtBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building account selection sql: %v: %w", err, rdb.ErrPrepare)
	}

	var account AccountRow
	var balance string
	if err = accountsView.rdb.QueryRow(sql, sqlArgs...).Scan(
		&account.Type,
		&account.Address,
		&account.Pubkey,
		&account.AccountNumber,
		&account.SequenceNumber,
		&balance,
	); err != nil {
		if errors.Is(err, rdb.ErrNoRows) {
			return nil, rdb.ErrNoRows
		}
		return nil, fmt.Errorf("error scanning account row: %v: %w", err, rdb.ErrQuery)
	}
	account.Balance = coin.MustParseCoinsNormalized(balance)
	return &account, nil
}

func (accountsView *Accounts) List(
	order AccountsListOrder,
	pagination *pagination.Pagination,
) ([]AccountRow, *pagination.PaginationResult, error) {
	stmtBuilder := accountsView.rdb.StmtBuilder.Select(
		"type",
		"address",
		"pubkey",
		"account_number",
		"sequence_number",
		"balance",
	).From(
		"view_accounts",
	)

	if order.AccountAddress == view.ORDER_DESC {
		stmtBuilder = stmtBuilder.OrderBy("address DESC")
	} else {
		stmtBuilder = stmtBuilder.OrderBy("address")
	}

	rDbPagination := rdb.NewRDbPaginationBuilder(
		pagination,
		accountsView.rdb,
	).BuildStmt(stmtBuilder)
	sql, sqlArgs, err := rDbPagination.ToStmtBuilder().ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("error building blocks select SQL: %v, %w", err, rdb.ErrBuildSQLStmt)
	}

	rowsResult, err := accountsView.rdb.Query(sql, sqlArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing blocks select SQL: %v: %w", err, rdb.ErrQuery)
	}

	accounts := make([]AccountRow, 0)
	for rowsResult.Next() {
		var account AccountRow
		var balance string
		if err = rowsResult.Scan(
			&account.Type,
			&account.Address,
			&account.Pubkey,
			&account.AccountNumber,
			&account.SequenceNumber,
			&balance,
		); err != nil {
			if errors.Is(err, rdb.ErrNoRows) {
				return nil, nil, rdb.ErrNoRows
			}
			return nil, nil, fmt.Errorf("error scanning account row: %v: %w", err, rdb.ErrQuery)
		}

		account.Balance = coin.MustParseCoinsNormalized(balance)
		accounts = append(accounts, account)
	}

	paginationResult, err := rDbPagination.Result()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing pagination result: %v", err)
	}

	return accounts, paginationResult, nil
}

type AccountRow struct {
	Type           string     `json:"type"`
	Address        string     `json:"address"`
	Pubkey         string     `json:"pubkey"`
	AccountNumber  string     `json:"accountNumber"`
	SequenceNumber string     `json:"sequenceNumber"`
	Balance        coin.Coins `json:"balance"`
}
