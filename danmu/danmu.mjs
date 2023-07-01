import { KeepLiveWS } from "bilibili-live-ws/browser";
const argv = process.argv;
const room = Number(argv[2]);
if (isNaN(room)) {
  throw new Error("room is required");
}
const live = new KeepLiveWS(room);
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
