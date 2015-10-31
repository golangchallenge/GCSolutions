```
# desktop
go run *.go

# android
gomobile build -target android .
adb install -r piano.apk
```

to support intended use on android (landscape, fullscreen), build with go mobile from
at least the following commit: 38a56c4998acb3d9e92277296a18e10302225654