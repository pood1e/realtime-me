import {
  ProviderConnectionAttemptStatus,
  ProviderConnectionStatus,
} from "@realtime-me/library-contracts";

export function terminalConnectionAttempt(status: ProviderConnectionAttemptStatus): boolean {
  return [
    ProviderConnectionAttemptStatus.CONNECTED,
    ProviderConnectionAttemptStatus.EXPIRED,
    ProviderConnectionAttemptStatus.REFUSED,
    ProviderConnectionAttemptStatus.FAILED,
  ].includes(status);
}

export function connectionAttemptStatus(status?: ProviderConnectionAttemptStatus): string {
  switch (status) {
    case ProviderConnectionAttemptStatus.SCANNED:
      return "已扫码，请在手机上确认";
    case ProviderConnectionAttemptStatus.CONNECTED:
      return "账号连接成功";
    case ProviderConnectionAttemptStatus.EXPIRED:
      return "二维码已过期，请关闭后重试";
    case ProviderConnectionAttemptStatus.REFUSED:
      return "登录已取消";
    case ProviderConnectionAttemptStatus.FAILED:
      return "登录失败，请关闭后重试";
    default:
      return "使用对应音乐 App 扫描二维码";
  }
}

export function providerConnectionDetail(status: ProviderConnectionStatus): string {
  if (status === ProviderConnectionStatus.NOT_CONFIGURED) return "需要先配置服务端应用凭据";
  if (status === ProviderConnectionStatus.RECONNECT_REQUIRED) return "登录已失效，请重新连接";
  return "尚未连接账号";
}
