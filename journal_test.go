package gtmcdc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Horolog2UnixTime(t *testing.T) {
	// run the following command in GT.M to check day values
	// write $$CDS^%H(day)

	// 1971/1/1 is 47117 days after 1840/12/31
	// any day that results in a date prior to 1971/1/1 in the
	// local timezone will return an error because
	// the timestamp value will be negative
	// Unix system cannot handle this sort of thing
	ts, err := Horolog2Timestamp("47117")
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ts)

	// this is a special case. GTM V6.3 sends this
	_, err = Horolog2Timestamp("0,0")
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ts)

	expected, _ := time.Parse("2006-01-02 15:04:05", "2019-09-26 16:35:00")
	ts, err = Horolog2Timestamp("65282,59700")
	assert.Nil(t, err)
	assert.Equal(t, expected.Unix(), ts)

	expected, _ = time.Parse("2006-01-02 15:04:05", "2019-10-04 00:24:45")
	ts, err = Horolog2Timestamp("65290,1485")
	assert.Nil(t, err)
	assert.Equal(t, expected.Unix(), ts)

	_, err = Horolog2Timestamp("29800130,1234")
	assert.NotNil(t, err)
}

func Test_Parse_JournalRecord_1(t *testing.T) {
	rec, _ := Parse(`04\0,0\78393877654\0\0\58808538812\0\0\0\0\^ZATFAMTBYPCT"`)
	fmt.Println(rec)
	// assert.Equal(t, "SET", rec.opcode)
	// assert.Equal(t, "300.00", rec.detail.value)

	// record is too short
	_, err := Parse(`05\65282,59700\28`)
	assert.NotNil(t, err)
}

func Test_Parse_NodeFlags(t *testing.T) {
	rec, _ := parseNodeFlags("SET", "^MSGLOG(\"))bCC#)(3\")")
	fmt.Println(rec[1])
}
func Test_Parse_JournalRecord_2(t *testing.T) {
	rec, _ := Parse(`08\65287,62154\3\0\0\3\0\0`)
	assert.Equal(t, "TSTART", rec.opcode)

	rec, _ = Parse(`09\65287,58606\8\0\0\8\0\0\1\`)
	assert.Equal(t, "TCOM", rec.opcode)
	assert.Equal(t, 8, rec.tran.tokenSeq)
	assert.Equal(t, "", rec.tran.tag)
	assert.Equal(t, "1", rec.tran.partners)
}

func Test_JournalRecord_Json(t *testing.T) {
	expected := `{"operand":"SET","transaction_num":"28",` +
		`"token_seq":28,"update_num":0,"stream_num":0,"stream_seq":0,` +
		`"journal_seq":0,"global":"ACN","key":"1234","subscripts":["51"],` +
		`"node_values":["300.00","61212","1","","","",""],` +
		`"time_stamp":1569515700}`

	rec, err := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^ACN(1234,51)="300.00|61212|1||||"`)
	assert.Nil(t, err)

	jayson, err := rec.JSON()

	assert.Nil(t, err)
	assert.Equal(t, expected, jayson)
}

func Test_atli(t *testing.T) {
	i := atoi("100")
	assert.Equal(t, 100, i)

	i = atoi("xx")
	assert.Equal(t, 0, i)
}
