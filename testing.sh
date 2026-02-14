#!/bin/sh

# check for the existence of the big-buck-bunny.mp4 and big-buck-bunny-frag.mp4

if ! test -e "resources/big-buck-bunny.mp4"; then 
    # curl the data
    curl -C - "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_5MB.mp4" -o "resources/big-buck-bunny.mp4"
fi

if ! test -e "resources/big-buck-bunny-frag.mp4"; then
    # ffmpeg
    ffmpeg -i "resources/big-buck-bunny.mp4" -c copy -movflags frag_keyframe+empty_moov "resources/big-buck-bunny-frag.mp4" 
fi

# once done, run the golang testing

echo "RUNNING THE TEST"
go test