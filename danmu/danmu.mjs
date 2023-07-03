import { KeepLiveWS } from "bilibili-live-ws";
import { encoder } from "bilibili-live-ws/src/buffer";
import { inflates } from "bilibili-live-ws/src/inflate/node";
const argv = process.argv;
const room = Number(argv[2]);
if (isNaN(room)) {
  throw new Error("room is required");
}
void (async function main() {
  const authBody = await getAuthBody(room);

  const live = new KeepLiveWS(room, { authBody: authBody });
  live.on("DANMU_MSG", (e) => {
    /**@type {[]string} */
    let info = e.info;
    if (info.length < 3) {
      return;
    }
    let danmu = info[1];
    /**@type {[]string} */
    let userInfo = info[2];
    if (userInfo.length < 2) {
      return;
    }
    console.log(`${userInfo[0]}|${danmu}`);
  });
})();

async function getAuthBody(roomid) {
  const bilipage = argv[3];
  if (!(typeof bilipage === "string" && bilipage !== "")) {
    return null;
  }
  const key = await fetch(
    new URL(
      `/live/xlive/web-room/v1/index/getDanmuInfo?id=${roomid}&type=0`,
      bilipage
    ),
    {
      headers: {
        "Js.fetch.credentials": "include",
      },
    }
  )
    .then((r) => r.json())
    .then((r) => {
      if (r.code !== 0) {
        throw r;
      }
      return r.data.token;
    });

  const [uid, buvid] = await fetch(new URL("/xhe-eval", bilipage), {
    method: "POST",
    body: `resolve(Promise.all(["DedeUserID","buvid3"].map(k=>cookieStore.get(k).then(v=>v.value))))`,
  }).then((r) => r.json());

  let auth = {
    uid: Number(uid),
    roomid: room,
    protover: 3,
    buvid: buvid,
    platform: "web",
    type: 2,
    key: key,
  };

  console.error("auth info got");

  return encoder("join", inflates, auth);
}
