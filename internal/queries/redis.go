package queries

import (
	"context"
    "encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/redhatinsights/payload-tracker-go/internal/structs"
)

func RetrieveRequestIdPayloadsFromRedis(rdb *redis.Client) RetrieveRequestIdPayloads {
	return func(dbQuery *gorm.DB, reqID string, sortBy string, sortDir string, verbosity string) []structs.PayloadData {
        /*
		val, err := rdb.Get(context.TODO(), reqID).Result()
		fmt.Println("err:", err)
        */
        val, err := rdb.ZRangeWithScores(context.TODO(), reqID, 0, 10).Result()
        if err != nil {
            fmt.Println("ZRange returned error:", err)
            return nil
        }

		fmt.Println("val:", val)

        payloadData := make([]structs.PayloadData, len(val))

        for i, zItem := range val {
            fmt.Println("zItem:", zItem)
            fmt.Printf("zItem:%T\n", zItem)

            s, ok := zItem.Member.(string)
            if !ok {
                fmt.Println("Cannot parse member")
            } else {
                fmt.Println("s:", s)
            }

            err := json.Unmarshal([]byte(s), &payloadData[i])
            if err != nil {
                fmt.Println("Cannot parse member")
            }

/*
            payloadData[i].OrgID = zItem.Member["org_id"].(string)
            payloadData[i].Service = zItem.Member["service"].(string)
            payloadData[i].Status = zItem.Member["status"].(string)
            payloadData[i].Date = time.Parse(time.RFC3339, zItem.Member["date"].(string))
*/
        }

        //zItem: {1.663969562032e+12 {"date":"2022-09-23T21:46:02.032661+00:00","org_id":"456","service":"advisor","status":"processing"}}

        return payloadData
	}
}

func RetrieveRequestIdPayloadsFromRedisFallbackToDB(rdb *redis.Client, retrieveFromDB RetrieveRequestIdPayloads) RetrieveRequestIdPayloads {

	retrieveFromRedis := RetrieveRequestIdPayloadsFromRedis(rdb)

	return func(dbQuery *gorm.DB, reqID string, sortBy string, sortDir string, verbosity string) []structs.PayloadData {

		results := retrieveFromRedis(dbQuery, reqID, sortBy, sortDir, verbosity)
		if len(results) > 0 {
			// FIXME:  Might need to go ahead and query db as well
			return results
		}

		return retrieveFromDB(dbQuery, reqID, sortBy, sortDir, verbosity)
	}
}
