# video stream recorder
live video stream recorder / vod service provider

downloader:
```
./vsr --url http://full/url/with/m3u8 --tail 24

url = m3u8 url to process
tail = how many hours keep
```

watch recorder stream via your favorite video player:
```
http://youripaddress:8080/start/<timestamp>/300/stream.m3u8
```
replace "{timestamp}" to localtime unix timestamp (https://www.epochconverter.com/)
replace 300 with how long video is (if you dont know, keep 300)
