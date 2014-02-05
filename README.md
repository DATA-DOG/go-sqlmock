
db = mock.Open("test", "")

db.ExpectTransactionBegin()
db.ExpectTransactionBegin().WillReturnError("some error")
db.ExpectQuery("SELECT bla").With(5, 8, "stat").WillReturnNone()
db.ExpectExec("UPDATE tbl SET").With(5, "val").WillReturnResult(res /* sql.Result */)
db.ExpectExec("INSERT INTO bla").With(5, 8, "stat").WillReturnResult(res /* sql.Result */)
db.ExpectQuery("SELECT bla").With(5, 8, "stat").WillReturnRows()

