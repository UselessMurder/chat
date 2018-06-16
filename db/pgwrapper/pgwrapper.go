package pgwrapper

import (
	"../pgconfig"
	"bufio"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"io"
	"os"
	"strings"
	"sync"
)

var db *sql.DB
var Wrapper DataBaseWrapper

type DataBaseWrapper struct {
	rlock         *sync.RWMutex
	connect_limit chan bool
	requestsList  map[string]*sql.Stmt
}

type transact struct {
	Tx   *sql.Tx
	Roll bool
}

func (dbw *DataBaseWrapper) initWrapper() {

	dbw.rlock = &sync.RWMutex{}
	dbw.requestsList = make(map[string]*sql.Stmt)
	dbw.connect_limit = make(chan bool, pgconfig.Config.Postgre_connect_count)
}

func (dbw *DataBaseWrapper) connect_up() {
	dbw.connect_limit <- true
}

func (dbw *DataBaseWrapper) connect_down() {
	<-dbw.connect_limit
}

func (dbw *DataBaseWrapper) setRequestList(file io.Reader) error {

	var err error

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		request := strings.Split(line, ":")
		if len(request) < 2 {
			return errors.New("Incorrect line.")
		}
		_, ok := dbw.requestsList[request[0]]
		if ok == true {
			err := dbw.requestsList[request[0]].Close()
			if err != nil {
				return err
			}
			delete(dbw.requestsList, request[0])
		}
		dbw.requestsList[request[0]], err = db.Prepare(request[1])
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (dbw *DataBaseWrapper) clearRequestList() error {

	for key := range dbw.requestsList {
		err := dbw.requestsList[key].Close()
		if err != nil {
			return err
		}
		delete(dbw.requestsList, key)
	}
	return nil
}

func (dbw *DataBaseWrapper) ReplaceRequestList(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	err = dbw.clearRequestList()
	if err != nil {
		return err
	}

	err = dbw.setRequestList(file)
	if err != nil {
		return err
	}

	return nil
}

func (dbw *DataBaseWrapper) UpdateRequestList(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	err = dbw.setRequestList(file)
	if err != nil {
		return err
	}

	return nil
}

func (dbw *DataBaseWrapper) CloseWrapper() error {

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	err := dbw.clearRequestList()
	if err != nil {
		return err
	}
	return nil
}

func (dbw *DataBaseWrapper) InsertRequest(name, value string) error {

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	var err error

	_, ok := dbw.requestsList[name]

	if ok == true {
		return errors.New("Request already exists!")
	}

	dbw.requestsList[name], err = db.Prepare(value)
	if err != nil {
		return err
	}

	return nil
}

func (dbw *DataBaseWrapper) UpdateRequest(name, value string) error {

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	_, ok := dbw.requestsList[name]

	if ok == false {
		return errors.New("Request does not exist!")
	}

	err := dbw.requestsList[name].Close()
	if err != nil {
		return err
	}

	delete(dbw.requestsList, name)

	dbw.requestsList[name], err = db.Prepare(value)
	if err != nil {
		return err
	}

	return nil
}

func (dbw *DataBaseWrapper) RemoveRequest(name, value string) error {

	dbw.rlock.Lock()
	defer dbw.rlock.Unlock()

	dbw.connect_up()
	defer dbw.connect_down()

	_, ok := dbw.requestsList[name]

	if ok == false {
		return errors.New("Request does not exist!")
	}

	err := dbw.requestsList[name].Close()
	if err != nil {
		return err
	}

	delete(dbw.requestsList, name)

	return nil
}

func (dbw *DataBaseWrapper) ExecTransact(requestName string, values ...interface{}) error {

	dbw.rlock.RLock()
	defer dbw.rlock.RUnlock()
	dbw.connect_up()
	defer dbw.connect_down()
	_, ok := dbw.requestsList[requestName]
	if !ok {
		return errors.New("Missmatch request!")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Stmt(dbw.requestsList[requestName]).Exec(values...)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbw *DataBaseWrapper) PrepareTransact(requestName string, values ...interface{}) (transact, error) {

	dbw.rlock.RLock()
	dbw.connect_up()

	var tx transact

	tx.Roll = true

	var err error

	_, ok := dbw.requestsList[requestName]
	if !ok {
		dbw.rlock.Unlock()
		dbw.connect_down()
		return tx, errors.New("Missmatch request!")
	}
	tx.Tx, err = db.Begin()
	if err != nil {
		dbw.rlock.Unlock()
		dbw.connect_down()
		return tx, err
	}
	_, err = tx.Tx.Stmt(dbw.requestsList[requestName]).Exec(values...)
	if err != nil {
		dbw.rlock.Unlock()
		dbw.connect_down()
		return tx, err
	}

	return tx, nil
}

func (dbw *DataBaseWrapper) RollBackTransact(tx transact) error {

	var err error
	if tx.Roll {
		err = tx.Tx.Rollback()
		dbw.rlock.Unlock()
		dbw.connect_down()
	}
	return err
}

func (dbw *DataBaseWrapper) CommitTransact(tx transact) error {

	err := tx.Tx.Commit()
	if err == nil {
		tx.Roll = false
	}
	dbw.rlock.Unlock()
	dbw.connect_down()
	return err
}

func (dbw *DataBaseWrapper) QueryRow(requestName string, values ...interface{}) (*sql.Row, error) {

	dbw.rlock.RLock()
	defer dbw.rlock.RUnlock()
	dbw.connect_up()
	defer dbw.connect_down()
	_, ok := dbw.requestsList[requestName]
	if !ok {
		return nil, errors.New("Missmatch request!")
	}

	row := dbw.requestsList[requestName].QueryRow(values...)

	return row, nil
}

func (dbw *DataBaseWrapper) Query(requestName string, values ...interface{}) (*sql.Rows, error) {

	dbw.rlock.RLock()
	defer dbw.rlock.RUnlock()
	dbw.connect_up()
	defer dbw.connect_down()
	_, ok := dbw.requestsList[requestName]
	if !ok {
		return nil, errors.New("Missmatch request!")
	}

	rows, err := dbw.requestsList[requestName].Query(values...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func init() {

	var err error

	db, err = sql.Open("postgres", "postgres://"+pgconfig.Config.Postgre_user+":"+pgconfig.Config.Postgre_password+"@"+pgconfig.Config.Postgre_host+"/"+pgconfig.Config.Postgre_database+"?sslmode="+pgconfig.Config.Postgre_ssl)
	if err != nil {
		panic("Postgresql writer not found!" + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("Postgresql not reply!:" + err.Error())
	}

	Wrapper.initWrapper()
}
