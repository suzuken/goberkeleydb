package goberkeleydb

import (
	"unsafe"
)

/*
 #cgo LDFLAGS: -ldb
 #include <stdlib.h>
 #include <db.h>
 static inline int db_open(DB *db, DB_TXN *txn, const char *file, const char *database, DBTYPE type, u_int32_t flags, int mode) {
 	return db->open(db, txn, file, database, type, flags, mode);
 }
 static inline int db_close(DB *db, u_int32_t flags) {
 	return db->close(db, flags);
 }
 static inline int db_get_type(DB *db, DBTYPE *type) {
 	return db->get_type(db, type);
 }
 static inline int db_put(DB *db, DB_TXN *txn, DBT *key, DBT *data, u_int32_t flags) {
 	return db->put(db, txn, key, data, flags);
 }
 static inline int db_get(DB *db, DB_TXN *txn, DBT *key, DBT *data, u_int32_t flags) {
 	return db->get(db, txn, key, data, flags);
 }
 static inline int db_del(DB *db, DB_TXN *txn, DBT *key, u_int32_t flags) {
 	return db->del(db, txn, key, flags);
 }
 static inline int db_cursor(DB *db, DB_TXN *txn, DBC **cursor, u_int32_t flags) {
 	return db->cursor(db, txn, cursor, flags);
 }
 static inline int db_cursor_close(DBC *cur) {
 	return cur->close(cur);
 }
 static inline int db_cursor_get(DBC *cur, DBT *key, DBT *data, u_int32_t flags) {
 	return cur->get(cur, key, data, flags);
 }

 static inline int db_env_open(DB_ENV *env, const char *home, u_int32_t flags, int mode) {
 	return env->open(env, home, flags, mode);
 }
 static inline int db_env_close(DB_ENV *env, u_int32_t flags) {
 	return env->close(env, flags);
 }

 static inline int db_env_txn_begin(DB_ENV *env, DB_TXN *parent, DB_TXN **txn, u_int32_t flags) {
 	return env->txn_begin(env, parent, txn, flags);
 }
 static inline int db_txn_abort(DB_TXN *txn) {
 	return txn->abort(txn);
 }
 static inline int db_txn_commit(DB_TXN *txn, u_int32_t flags) {
 	return txn->commit(txn, flags);
 }
*/
import "C"

type BerkeleyDB struct {
	ptr *C.DB
}
type Transaction struct {
	ptr *C.DB_TXN
}
type Environment struct {
	ptr *C.DB_ENV
}
type DBT struct {
	ptr *C.DBT
}
type Cursor struct {
	ptr *C.DBC
}

var NoEnv = Environment{ptr: nil}
var NoTxn = Transaction{ptr: nil}

type DbType int

// Available database types.
const (
	BTree    = DbType(C.DB_BTREE)
	Hash     = DbType(C.DB_HASH)
	Numbered = DbType(C.DB_RECNO)
	Queue    = DbType(C.DB_QUEUE)
	Unknown  = DbType(C.DB_UNKNOWN)
)

type DbFlag uint32

const (
	DB_AUTO_COMMIT      = DbFlag(C.DB_AUTO_COMMIT)
	DB_CREATE           = DbFlag(C.DB_CREATE)
	DB_EXCL             = DbFlag(C.DB_EXCL)
	DB_MULTIVERSION     = DbFlag(C.DB_MULTIVERSION)
	DB_NOMMAP           = DbFlag(C.DB_NOMMAP)
	DB_RDONLY           = DbFlag(C.DB_RDONLY)
	DB_READ_UNCOMMITTED = DbFlag(C.DB_READ_UNCOMMITTED)
	DB_THREAD           = DbFlag(C.DB_THREAD)
	DB_TRUNCATE         = DbFlag(C.DB_TRUNCATE)
)

func Err(C.int) error {
	return nil
}

func bytesDBT(val []byte) *C.DBT {
	return &C.DBT{
		data:  unsafe.Pointer(&val[0]),
		size:  C.u_int32_t(len(val)),
		flags: C.DB_DBT_READONLY,
	}
}

func cloneToBytes(val *C.DBT) []byte {
	return C.GoBytes(val.data, C.int(val.size))
}

func OpenBDB(env Environment, txn Transaction, file string, database *string, dbtype DbType, flags DbFlag, mode int) (*BerkeleyDB, error) {
	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))

	var cDatabase *C.char
	if database != nil {
		cDatabase = C.CString(*database)
		defer C.free(unsafe.Pointer(cDatabase))
	}

	var db *BerkeleyDB = new(BerkeleyDB)

	//The flags parameter is currently unused, and must be set to 0.
	// https://docs.oracle.com/cd/E17276_01/html/api_reference/C/dbcreate.html
	err := Err(C.db_create(&db.ptr, env.ptr, 0))
	if err != nil {
		return nil, err
	}

	err = Err(C.db_open(db.ptr, txn.ptr, cFile, cDatabase, C.DBTYPE(dbtype), C.u_int32_t(flags), C.int(mode)))
	if err != nil {
		db.Close(0)
		return nil, err
	}
	return db, nil
}
func (db BerkeleyDB) Close(flags DbFlag) (err error) {
	if db.ptr != nil {
		err = Err(C.db_close(db.ptr, C.u_int32_t(flags)))
		db.ptr = nil
	}
	return err
}
func (db BerkeleyDB) GetType() (DbType, error) {
	var cdbtype C.DBTYPE
	err := Err(C.db_get_type(db.ptr, &cdbtype))
	dbtype := DbType(cdbtype)
	return dbtype, err
}
func (db BerkeleyDB) Put(txn Transaction, key, val []byte, flags DbFlag) error {
	return Err(C.db_put(db.ptr, txn.ptr, bytesDBT(key), bytesDBT(val), C.u_int32_t(flags)))
}
func (db BerkeleyDB) Get(txn Transaction, key []byte, flags DbFlag) ([]byte, error) {
	var data C.DBT
	data.flags |= C.DB_DBT_REALLOC
	defer C.free(data.data)

	err := Err(C.db_get(db.ptr, txn.ptr, bytesDBT(key), &data, C.u_int32_t(flags)))
	if err != nil {
		return nil, err
	}

	return cloneToBytes(&data), nil
}
func (db BerkeleyDB) Del(txn Transaction, key []byte, flags DbFlag) error {
	return Err(C.db_del(db.ptr, txn.ptr, bytesDBT(key), C.u_int32_t(flags)))
}

func (db BerkeleyDB) NewCursor(txn Transaction, flags DbFlag) (*Cursor, error) {
	ret := new(Cursor)
	err := Err(C.db_cursor(db.ptr, txn.ptr, &ret.ptr, C.u_int32_t(flags)))
	if err != nil {
		return nil, err
	}
	return ret, nil
}
func (cursor Cursor) Close() error {
	err := Err(C.db_cursor_close(cursor.ptr))
	if err == nil {
		cursor.ptr = nil
	}
	return err
}
func (cursor Cursor) First() ([]byte, []byte, error) {
	return cursor.CursorGetRaw(C.DB_FIRST)
}
func (cursor Cursor) Next() ([]byte, []byte, error) {
	return cursor.CursorGetRaw(C.DB_NEXT)
}
func (cursor Cursor) Last() ([]byte, []byte, error) {
	return cursor.CursorGetRaw(C.DB_LAST)
}
func (cursor Cursor) CursorGetRaw(flags DbFlag) ([]byte, []byte, error) {
	key := C.DBT{flags: C.DB_DBT_REALLOC}
	defer C.free(key.data)
	val := C.DBT{flags: C.DB_DBT_REALLOC}
	defer C.free(val.data)

	err := Err(C.db_cursor_get(cursor.ptr, &key, &val, C.u_int32_t(flags)))
	if err != nil {
		return nil, nil, err
	}
	return cloneToBytes(&key), cloneToBytes(&val), nil
}
func NewEnvironment(home string, flags DbFlag, mode int) (*Environment, error) {
	return nil, nil
}
func (env Environment) Close(flags DbFlag) error {
	return nil
}

func (env Environment) BeginTransaction(flags DbFlag) (*Transaction, error) {
	return nil, nil
}
func (trx Transaction) Abort() error {
	return nil
}
func (txn Transaction) Commit(flags DbFlag) error {
	return nil
}
