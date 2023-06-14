docker buildx build -t local/mmfm-playback-go --platform=linux/arm/v7 . --load
id = $(docker create "local/mmfm-playback-go")
docker cp $id:/app/mmfm-go-armv7 ./mmfm-go-armv7-v2
docker rm $id