# This app parse HTML documents from google images

For run this program:
```
go run main.go -text [img text that you want find]
```
Or build it and run:
```
go build main.go
```
# How to get help indo?
For more flags info just enter this line
```
go run main.go -?
```
```
./main -?
```
# Available flags
-text flags is required to start app

Others flags are optional
```
 -graceful-timeout duration
        the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m (default 15s)
  -img-storage string
        Path for saving scrapped images links (default "storage")
  -resp-log string
        Path for saving responses in txt format (default "log")
  -tbs string

                Set the period for which the image was published
                If you want imagies in all period just don't enter this flag
                Params:
                        d - in 24 hours period
                        w - in week period
                        m - in month period
                        y - in year period

  -text string
        That will be serched in yandex images
```
