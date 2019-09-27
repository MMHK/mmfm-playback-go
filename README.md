# mmfm-playback-go

为 `MMFM` 提供后台播放器。

此项目是对`ffplay`封装执行播放任务，并通过`websocket`([socket.io](https://socket.io)) 与 `MMFM` 通讯同步播放器状态。

用于嵌入式系统或者无桌面`linux`系统的音频播放。

`ffmpeg`的可执行文件可以在[这里](https://ffbinaries.com/downloads)下载，
不过建议linux系统直接 `yum` 及 `apt` 安装。

请不要尝试在 `docker` 运行，linux下的声音驱动是个深坑。

## 项目依赖

- [socket.io](https://socket.io)，成熟的 `websocket` 多端通讯方案。
- [ffmpeg 4.x](https://www.ffmpeg.org/)，成熟的流媒体播放及编码/解码方案。

## 编译项目

```bash
# 该项目已经支持go mod
export GO111MODULE=on
go mod vendor
go build -o mmfm-playback-go .
```

## 执行

```bash
mmfm-playback-go -c ./conf.json
```

## 配置文件说明

```json
{
    "ffmpeg": {
        "ffplay": "...", 
        "ffprobe": "..."
    },
    "ws": "ws://localhost:8888/io/?EIO=3\u0026transport=websocket",
    "cache": "...",
    "web": "http://localhost:8888/song/get"
}
```

|key|说明|
|-|-|
|ffmpeg.ffplay|ffplay 执行文件位置，linux下使用 which ffplay获取|
|ffmpeg.ffprobe|ffprobe 执行文件位置，linux下使用 which ffprobe获取|
|ws|`mmfm` websocket 通讯地址|
|cache|音频文件缓存位置，建议使用系统临时目录，重新即烧毁|
|web|`mmfm` 获取歌曲地址api|


## 开发日志

- `ffmpeg`， 已经将原来项目依赖的 `Player`，更换为`ffmpeg` 使用cli wrapper的方式进行音乐的播放。
  因为QQ音乐的API返回的是`m4a`格式，这种格式属于`mp4`的一个子集，所以更换成支持格式更广泛的`ffmpeg`。

  安装之前请确保系统中已经安装好 `ffmpeg`。

- `Linux` 下 `ALSA` 下的问题 , 虽然 `ALSA` 包含了大部分的声卡驱动， but 这linux下的驱动有点麻烦，驱动是通用的，
   还需要配合不同的配置参数，才能正常输出声音的。具体步骤：
   
   - 使用 `aplay -l` / `alasmixer` ，确保声音模块已经正常加载。
   - 如果发现自己的声卡型号带有 `HDA` 开头的话，都建议使用 `snd-hda-intel` 这个内核驱动。
   - 使用 `cat /proc/asound/card0/codec* | grep Codec` ，确认一下所属的声卡芯片型号。
   - 去这个[页面](http://lxr.linux.no/linux+v3.2.19/Documentation/sound/alsa/HD-Audio-Models.txt) 找 `ALSA` 支持的芯片模式。
   - 修改 `/etc/modprobe.d/alsa-base.conf` 文件。
   - 加入 `options snd-hda-intel model=[model_name]` ， 后面的 `model_name`, 请根据 [HD-Audio-Models.txt](http://lxr.linux.no/linux+v3.2.19/Documentation/sound/alsa/HD-Audio-Models.txt)， 里面的芯片模式逐个尝试。
   - 修改完 `alsa-base.conf`, 使用 `alsa force-reload` 激活配置，如果幸运的话，试几个model之后就能出声音了...
   
