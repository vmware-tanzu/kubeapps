export default class WebSocketHelper {
  public static apiBase() {
    let apiBase: string;

    // Use WebSockets Secure if using HTTPS and WebSockets if not
    if (location.protocol === "https:") {
      apiBase = `wss://${window.location.host}${window.location.pathname}api/kube`;
    } else {
      apiBase = `ws://${window.location.host}${window.location.pathname}api/kube`;
    }
    return apiBase;
  }
}
