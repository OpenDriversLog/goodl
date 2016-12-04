// Package dataConverter is responsible for converting different formats to ODL-compatible CSV
package dataConverter

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	//pretty "github.com/tonnerre/golang-pretty"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Compufreak345/dbg"
	. "github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/webfw"
)

const TAG = "goodl/dataConverter.go"

var Err_UnknownFormat = errors.New("Unknown format")
var Err_NoData = errors.New("No data")

// ConvertAnythingToCSV converts any given string to a CSV, if the given format is known to us.
// returns CSV and Err_UnknownFormat, Err_NoData or other / nil error if anything fails / not fails.
// all data with timeStamp < getDataSince will be ignored.
func ConvertAnythingToCSV(src string, format string, getDataSince int64) (res string, err error) {
	switch format {
	case "NMEA/GPRMC":
		return ConvertNMEAToCSV(src, getDataSince)
	case "KML":
		return ConvertKMLToCSV(src, getDataSince)
	default:
		return "", Err_UnknownFormat
	}
}

// ConvertKMLToCSV converts a KML-file to a CSV. Anything before getDataSince will be ignored.
func ConvertKMLToCSV(src string, getDataSince int64) (res string, err error) {
	dbg.I(TAG, "ConvertKMLToCSV since ", getDataSince)
	if src == "" {
		return "", Err_NoData
	}
	var buf bytes.Buffer

	buf.WriteString("timeMillis,latitude,longitude,altitude,accuracy,provider,source,accuracyRating,speed")

	XMLdata := []byte(src)

	//KML Schema is defined in goodl-lib/models/KMLDTOs.go
	q := KML{}
	//now parsing XML into given structs, can be empty
	xml_err := xml.Unmarshal(XMLdata, &q)

	//pretty.Print("q: \n  ", q)

	if xml_err != nil {
		dbg.E("xml_err: ", "%v", xml_err)
		return
	}

	// in case of error in structure first DP will be empty
	if q.Document.Placemark.GxTrack == nil || q.Document.Placemark.GxTrack[0].When == nil {
		dbg.E("xml_err: ", "%v", xml_err)
		return "", Err_NoData
	}

	for index, curWhen := range q.Document.Placemark.GxTrack[0].When {
		curCords := strings.Split(q.Document.Placemark.GxTrack[0].Coord[index], " ")
		curCordsX, err := strconv.ParseFloat(curCords[1], 64)
		if err != nil {
			dbg.I(TAG, "Could not parse long at line ", index)
			return "", Err_UnknownFormat
		}

		curCordsY, err := strconv.ParseFloat(curCords[0], 64)
		if err != nil {
			dbg.I(TAG, "Could not parse lat at line ", index)
			return "", Err_UnknownFormat
		}
		curTime, err := time.Parse(time.RFC3339, curWhen)
		if err != nil {
			dbg.I(TAG, "Could not parse time at line ", index)
			return "", Err_UnknownFormat
		}

		//TODO Check TimeZone foo
		timeMillis := time.Date(curTime.Year(), curTime.Month(), curTime.Day(), curTime.Hour(), curTime.Minute(), curTime.Second(), 0, webfw.Config().TimeConfig.TimeLocation).UnixNano() / 1000 / 1000
		if timeMillis >= getDataSince {
			buf.WriteString(fmt.Sprintf("\r\n%d,%f,%f,%d,%d,%s,%d,%d,%f", timeMillis, curCordsX, curCordsY, 0, 0, "GPS", 0, 0, 0))
		}
	}

	return buf.String(), nil
}

// ConvertNMEAToCSV converts NMEA to CSV - currently supporting ONLY $GPRMC (Recommended minimum specific GPS/Transit data).
// Anything before getDataSince will be ignored.
func ConvertNMEAToCSV(src string, getDataSince int64) (res string, err error) {
	if src == "" {
		return "", Err_NoData
	}
	var buf bytes.Buffer
	buf.WriteString("timeMillis,latitude,longitude,altitude,accuracy,provider,source,accuracyRating,speed")

	// mock first line to push reader to always use 13 columns and throw errors if otherwise
	rd := csv.NewReader(strings.NewReader(",,,,,,,,,,,,\r\n" + src))
	_, _ = rd.Read()
	recCnt := 0
	firstTry := true
	for {
		recCnt++
		var record []string
		record, err = rd.Read()

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			if strings.Contains(err.Error(), csv.ErrFieldCount.Error()) {
				if(firstTry) {
					firstTry = false
					// try with 12 instead of 13 columns - no clue why, but first prototype writes 12 instead of 13 columns
					rd = csv.NewReader(strings.NewReader(",,,,,,,,,,,\r\n" + src))
					_, _ = rd.Read()
					recCnt = 0
				}
				dbg.I(TAG, "Row %d broken - wrong length of %d (%s)", recCnt, len(record), err)
				continue
			}
			dbg.E(TAG, "Failed to read CSV : %v", err)
			return
		}

		rtype := record[0]
		if rtype != "$GPRMC" {
			return "", Err_UnknownFormat
		}
		if record[2] == "V" { // Warning - probably no GPS signal - skip this
			continue

		}

		var timeStamp = record[1]
		var date = record[9]

		var northSouth = record[4]
		var eathWest = record[6]

		nSCoordH, err := strconv.ParseFloat(record[3][0:2], 64)
		nSCoordM, err := strconv.ParseFloat(record[3][2:], 64)
		if err != nil { // TODO: Add more detailed error for user
			dbg.I(TAG, "Could not parse nSCoord at line ", recCnt)
			return "", Err_UnknownFormat
		}
		eWCoordH, err := strconv.ParseFloat(record[5][0:3], 64)
		eWCoordM, err := strconv.ParseFloat(record[5][3:], 64)

		if err != nil {
			dbg.I(TAG, "Could not parse eWCoord at line ", recCnt)
			return "", Err_UnknownFormat
		}
		nSCoord := nSCoordH + nSCoordM/60
		eWCoord := eWCoordH + eWCoordM/60

		if northSouth == "S" {
			nSCoord = -1 * nSCoord
		}
		if eathWest == "W" {
			eWCoord = -1 * eWCoord
		}
		speed, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			dbg.I(TAG, "Could not parse speed at line ", recCnt)
			return "", Err_UnknownFormat
		}

		speed = KnotsToKmh(speed)
		year, err := strconv.Atoi("20" + date[4:6])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		month, err := strconv.Atoi(date[2:4])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		day, err := strconv.Atoi(date[0:2])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		hour, err := strconv.Atoi(timeStamp[0:2])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		mins, err := strconv.Atoi(timeStamp[2:4])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		secs, err := strconv.Atoi(timeStamp[4:6])
		if err != nil {
			dbg.I(TAG, "Could not parse time/date at line ", recCnt)
			return "", Err_UnknownFormat
		}
		timeMillis := time.Date(year, time.Month(month), day, hour, mins, secs, 0, webfw.Config().TimeConfig.TimeLocation).UnixNano() / 1000 / 1000

		if timeMillis >= getDataSince {
			buf.WriteString(fmt.Sprintf("\r\n%d,%f,%f,%d,%d,%s,%d,%d,%f", timeMillis, nSCoord, eWCoord, 0, 0, "GPS", 0, 0, speed))
		}
	}

	return buf.String(), nil
}

// KnotsToKmh converts Knots to kmh
func KnotsToKmh(knots float64) (kmh float64) {
	return knots * 1.852
}
