package model

import (
	"database/sql"
	"fmt"
	"time"
)

type namedQuery struct {
	ID       int64
	Name     string
	Query    string
	Added    NullTime
	Modified NullTime
}

func (model *GrogModel) getNamedQueries() ([]*namedQuery, error) {
	var foundQueries []*namedQuery
	var err error

	rows, rowsErr := model.db.DB.Query("select id, name, query, added, modified from queries")

	if rowsErr == nil {
		defer rows.Close()

		var (
			id     int64
			qname  string
			query  string
			added  int64
			edited int64
		)

		foundQueries = make([]*namedQuery, 0, 10) // 10 is arbitrary

		for rows.Next() {
			if rows.Scan(&id, &qname, &query, &added, &edited) != sql.ErrNoRows {
				foundQuery := new(namedQuery)
				foundQuery.ID = id
				foundQuery.Name = qname
				foundQuery.Query = query
				foundQuery.Added.Set(time.Unix(added, 0))
				foundQuery.Modified.Set(time.Unix(edited, 0))

				foundQueries = append(foundQueries, foundQuery)
			}
		}

	} else {
		err = rowsErr
	}

	return foundQueries, err
}

// NamedQueryFunc is the signature of a function that can be called to execute a query
// that is stored in the database.
type NamedQueryFunc func([]interface{}) ([]map[string]string, error)

// NamedQueries stores NamedQueryFuncs according to the name of the namedquery.
type NamedQueries map[string]NamedQueryFunc

// MakeNamedQuerier returns a function that can be called to execute a named query from the
// database.
func (model *GrogModel) MakeNamedQuerier(query *namedQuery) NamedQueryFunc {
	db := model.db.DB

	return func(params []interface{}) ([]map[string]string, error) {
		var queryResults []map[string]string
		var err error

		nqResults, nqErr := db.Query(query.Query, params...)

		if nqErr == nil {
			defer nqResults.Close()

			queryResults = make([]map[string]string, 0, 10) // Cap of 10 is arbitrary.
			columnNames, columnNamesErr := nqResults.Columns()

			if columnNamesErr != nil {
				panic(fmt.Sprintf("namedquery[%s]: error getting names of columns from database: %v", query.Name, columnNamesErr))
			}

			nqRowResults := make([]interface{}, len(columnNames))
			for i := range nqRowResults {
				nqRowResults[i] = new(string)
			}

			for nqResults.Next() {
				nqResErr := nqResults.Scan(nqRowResults...)

				if nqResErr != nil {
					panic(fmt.Sprintf("namedquery[%s]: error scanning results: %v", query.Name, nqResErr))
				}

				nqRowResultsMapped := make(map[string]string)
				for i, colName := range columnNames {
					nqRowResultsMapped[colName] = *nqRowResults[i].(*string)
				}

				queryResults = append(queryResults, nqRowResultsMapped)
			}
		} else {
			err = fmt.Errorf("database error while executing namedquery %s: %v", query.Name, nqErr)
		}

		return queryResults, err
	}
}

// LoadNamedQueries retries all namedquery rows and creates an invoker func for each.
func (model *GrogModel) LoadNamedQueries() map[string]NamedQueryFunc {
	allQueries, allQueriesErr := model.getNamedQueries()

	if allQueriesErr != nil {
		panic(fmt.Sprintf("error loading namedqueries: %v", allQueriesErr))
	}

	allQueryFuncs := make(map[string]NamedQueryFunc)

	for _, q := range allQueries {
		querierFunc := model.MakeNamedQuerier(q)
		allQueryFuncs[q.Name] = querierFunc
	}

	return allQueryFuncs
}
