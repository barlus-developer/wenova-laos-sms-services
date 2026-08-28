// Package wenovasms is a small, secure Go client for the Wenova Link
// SMS API. It sends SMS (OTP, notifications, promotions) to Lao mobile
// numbers in 20XXXXXXXX format.
//
// Create a client with New, then send messages with SendOTP, SendInfo,
// or SendCustom:
//
//	wnv, err := wenovasms.New(wenovasms.Config{Key: "your-api-token"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	res, err := wnv.SendOTP(context.Background(), "2012345678", "Your OTP is: 123456")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The API token is only placed in the request body and is never
// included in errors or logs.
package wenovasms
