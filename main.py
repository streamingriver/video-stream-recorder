#!/usr/bin/env python3

import sqlite3
import m3u8
import cacheout
import urllib.request
import sys
import os
import time
from bottle import Bottle, run, static_file
import threading
from optparse import OptionParser

try:
    os.umask(0)
    os.mkdir("./files", 0o777)
except:
    pass

parser = OptionParser()
parser.add_option("-u", "--url")
parser.add_option("-t", "--tail", default=24)
parser.add_option("-d", "--host", default="http://localhost:8080")
parser.add_option("-p", "--port", default=8080)

(options, args) = parser.parse_args()

if options.url is None:
    print("set flag --url http://stream.m3u8")
    sys.exit(1)


conn = sqlite3.connect("files.sqlite3")
conn.row_factory = sqlite3.Row

cursor = conn.cursor()

cursor.execute('''
create table if not exists files 
(downloadtime int, duration float, name text unique);
''')
conn.commit()

m3u8_str = "stream.m3u8"
url = options.url
parts = url.split("/")
m3u8_str = parts[len(parts)-1]
url = url.replace(m3u8_str, "%s")

cache = cacheout.FIFOCache()
cache.configure(maxsize=10)

app = Bottle()

@app.route('/start/<start>/<end>')
def greet(start,end):
    out = ""
    conn = sqlite3.connect("files.sqlite3")
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    rows = cursor.execute("select * from files where downloadtime >= ? and downloadtime <= ?", (start, time.time()+(int(end)*60)))
    out += """#EXTM3U
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-TARGETDURATION:50
#EXT-X-VERSION:4
#EXT-X-MEDIA-SEQUENCE:0"""
    for row in rows:
        out += "#EXTINF:%s"%row["duration"]+"\n"
        out += options.host+"/play/"+row["name"]+"\n"
    out += ("#EXT-X-ENDLIST")
    conn.close()
    return out

@app.route('/play/<file>')
def play(file):
    return static_file(file, root="./files")

def thread0():
    run(app, host='0.0.0.0', port=options.port)

def thread1():
    while True: 
        time.sleep(1)

        m = m3u8.load(url % m3u8_str)

        for k in m.segments:
            if cache.get(k.uri) == None:
                # print("Should cache %s" % (url % k.uri))
                cache.set(k.uri, 1)
                urllib.request.urlretrieve(url % k.uri, "./files/"+k.uri)
                try:
                    conn = sqlite3.connect("files.sqlite3")
                    conn.row_factory = sqlite3.Row
                    conn.execute('insert into files values (?,?,?)', (time.time(), k.duration, k.uri))
                    conn.commit()
                except:
                    pass

def thread2():
    while True: 
        print("Deleting files older than %s hours" % options.tail)
        conn = sqlite3.connect("files.sqlite3")
        conn.row_factory = sqlite3.Row
        cursor = conn.cursor()
        dt = time.time()-int(options.tail)*60*60
        rows = cursor.execute("select * from files where downloadtime <= %d" % (dt))
        for row in rows:
            f = "./files/%s" % row["name"]
            try:
                os.remove(f)
                q = "delete from files where downloadtime = %s" % (row["downloadtime"])
                cursor.execute(q)
            except:
                pass
        conn.commit()
        conn.close()
        time.sleep(60)

if __name__ == "__main__":
    x = threading.Thread(target=thread1, daemon=True)
    x1 = threading.Thread(target=thread0, daemon=True)
    x2 = threading.Thread(target=thread2, daemon=True)
    x2.start()
    x1.start()
    x.start()
    x.join()
