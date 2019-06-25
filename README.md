# video stream recorder
live video stream recorder / vod service provider

downloader:
```
./vod -bind :8080 -db /mnt/channelname.db -ttl 24h -name channelname -root /mnt/disk1 -url http://your/hls/stream.m3u8
```

watch recorder stream via your favorite video player:
```
http://youripaddress:8080/channelname/timeshift_abs_mono-[timestamp].m3u8
```
replace "{timestamp}" to localtime unix timestamp (https://www.epochconverter.com/)
