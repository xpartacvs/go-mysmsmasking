# MySMSMasking

[![Go Reference](https://pkg.go.dev/badge/github.com/xpartacvs/go-mysmsmasking.svg)](https://pkg.go.dev/github.com/xpartacvs/go-mysmsmasking)

Package golang untuk berkomunikasi dengan API [MySMSMasking](https://mysmsmasking.com/)

## Cara Pakai

Download paket:

```bash
go get github.com/xpartacvs/go-mysmsmasking
```

>Sementara ini cuma ada satu package didalamnya yaitu package bernama `sms`.

Kemudian kita langsung koding:

```go
package main

import (
    "fmt"
    "github.com/xpartacvs/go-mysmsmasking/sms"
)

func main() {

    // Instansiasi client
    client := sms.NewClient("<username>", "<password>")

    // Cek saldo dan kedaluarsa
    acc, err := client.GetAccountInfo()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Saldo\t\t: %d\n", acc.Balance)
    fmt.Printf("Kedaluarsa\t: %s\n\n", acc.Expiry.Format("2006-01-02 15:04:05 MST"))

    // Kirim SMS
    awb, err := client.Send("081xxxxxxxxxx", "Testing kirim SMS dengan package github.com/xpartacvs/go-mysmsmasking")
    if err != nil {
        panic(err)
    }
    fmt.Printf("ID Kiriman\t: %s\n", awb.Id)
    fmt.Printf("Waktu Kirim\t: %s\n\n", awb.Timestamp.Format("2006-01-02 15:04:05 MST"))

    // Cek status kiriman SMS
    status, err := client.GetStatus(awb.Id)
    if err != nil {
        panic(err)
    }
    switch status {
    case sms.DELIVERED:
        fmt.Printf("Status ID %s: SAMPAI", awb.Id)
    case sms.SENT:
        fmt.Printf("Status ID %s: OTW", awb.Id)
    case sms.INVALID_ID:
        fmt.Printf("Status ID %s: ID TIDAK KETEMU", awb.Id)
    default:
        fmt.Printf("Status ID %s: GAGAL", awb.Id)
    }

}
```
