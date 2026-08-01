export default {
  brand: {
    name: 'ZlieA'
  },
  console: {
    title: 'ZlieA证书管理控制台'
  },
  common: {
    cancel: '取消',
    confirm: '确定',
    refresh: '刷新',
    operation: '操作',
    detail: '详情',
    download: '下载',
    submit: '提交',
    empty: '暂无数据'
  },
  route: {
    login: '登录',
    certUpload: '证书上传',
    certDownload: '证书下载'
  },
  lang: {
    label: '语言',
    zh: '中文',
    en: 'English'
  },
  login: {
    username: '用户名',
    password: '密码',
    placeholderUsername: '请输入用户名',
    placeholderPassword: '请输入密码',
    button: '登 录',
    success: '登录成功',
    failed: '登录失败，用户名或者密码错误',
    ruleUsername: '请输入用户名',
    rulePassword: '请输入密码'
  },
  layout: {
    logout: '退出登录',
    logoutButton: '退出',
    logoutConfirmTitle: '提示',
    logoutConfirmPrompt: '确定要退出登录吗？'
  },
  cert: {
    upload: {
      title: '证书上传',
      alias: '证书别名',
      placeholderAlias: '请选择证书',
      ruleAlias: '请选择证书别名',
      fullchain: '证书链',
      privkey: '私钥',
      fullchainTip: '完整证书链 PEM 文件（fullchain.pem）。',
      privkeyTip: '私钥 PEM 文件（privkey.pem）。',
      selectFile: '选择文件',
      submit: '提交上传',
      ruleFullchain: '请选择 fullchain 文件',
      rulePrivkey: '请选择 privkey 文件',
      warnMultipleFile: '每个位置只允许上传一个文件，请先删除已有文件',
      success: '证书上传成功',
      alertTitle: '当前没有可上传的证书',
      alertDesc:
        '仅 auto_renew 为 false 的证书允许手动上传；auto_renew 为 true 的证书由 ACME 自动签发与续期。'
    },
    download: {
      title: '证书下载',
      refresh: '刷新',
      empty: '暂无证书',
      columnAlias: '别名',
      columnDomain: '主域名',
      columnProvider: 'DNS 提供商',
      columnRenewMode: '续期模式',
      tagAuto: '自动',
      tagManual: '手动',
      columnNotBefore: '证书生效时间',
      columnNotAfter: '证书失效时间',
      columnDaysRemaining: '过期天数',
      columnExpectedRenewAt: '预计更新日期',
      tagExpired: '已过期',
      columnOperation: '操作',
      btnDetail: '详情',
      btnDownload: '下载',
      startDownload: '开始下载',
      invalidPackage: '下载失败：响应不是有效的证书包',
      detail: {
        title: '证书详情',
        alias: '别名',
        cn: '通用名 (CN)',
        san: 'SAN 域名',
        notBefore: '生效时间',
        notAfter: '过期时间',
        daysRemaining: '剩余天数',
        expired: '已过期',
        expectedRenewAt: '预计更新时间',
        empty: '暂无数据'
      }
    }
  },
  request: {
    failed: '请求失败',
    failedWithStatus: '请求失败 ({status})',
    expired: '登录已过期，请重新登录',
    network: '网络异常，请检查与服务端的连接',
    unknown: '未知错误'
  }
}
