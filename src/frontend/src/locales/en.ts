export default {
  brand: {
    name: 'ZlieA'
  },
  console: {
    title: 'ZlieA Certificate Management'
  },
  common: {
    cancel: 'Cancel',
    confirm: 'Confirm',
    refresh: 'Refresh',
    operation: 'Action',
    detail: 'Detail',
    download: 'Download',
    submit: 'Submit',
    empty: 'No data'
  },
  route: {
    login: 'Login',
    certUpload: 'Upload Certificate',
    certDownload: 'Download Certificate'
  },
  lang: {
    label: 'Language',
    zh: '中文',
    en: 'English'
  },
  login: {
    username: 'Username',
    password: 'Password',
    placeholderUsername: 'Enter your username',
    placeholderPassword: 'Enter your password',
    button: 'Sign In',
    success: 'Signed in successfully',
    failed: 'Sign-in failed: incorrect username or password',
    ruleUsername: 'Please enter your username',
    rulePassword: 'Please enter your password'
  },
  layout: {
    logout: 'Sign Out',
    logoutButton: 'Sign Out',
    logoutConfirmTitle: 'Notice',
    logoutConfirmPrompt: 'Are you sure you want to sign out?'
  },
  cert: {
    upload: {
      title: 'Upload Certificate',
      alias: 'Certificate alias',
      placeholderAlias: 'Select a certificate',
      ruleAlias: 'Please select a certificate alias',
      fullchain: 'Certificate chain',
      privkey: 'Private key',
      fullchainTip: 'Full certificate chain PEM file (fullchain.pem).',
      privkeyTip: 'Private key PEM file (privkey.pem).',
      selectFile: 'Choose file',
      submit: 'Upload',
      ruleFullchain: 'Please select the fullchain file',
      rulePrivkey: 'Please select the privkey file',
      warnMultipleFile: 'Only one file is allowed per slot. Please remove the existing file first.',
      success: 'Certificate uploaded successfully',
      alertTitle: 'No certificates available for upload',
      alertDesc:
        'Only certificates with auto_renew set to false can be uploaded manually. Certificates with auto_renew set to true are issued and renewed automatically via ACME.'
    },
    download: {
      title: 'Download Certificate',
      refresh: 'Refresh',
      empty: 'No certificates',
      columnAlias: 'Alias',
      columnDomain: 'Primary domain',
      columnProvider: 'DNS provider',
      columnRenewMode: 'Renewal',
      tagAuto: 'Auto',
      tagManual: 'Manual',
      columnNotBefore: 'Valid from',
      columnNotAfter: 'Expires',
      columnDaysRemaining: 'Days remaining',
      columnExpectedRenewAt: 'Expected renewal',
      tagExpired: 'Expired',
      columnOperation: 'Action',
      btnDetail: 'Detail',
      btnDownload: 'Download',
      startDownload: 'Download started',
      invalidPackage: 'Download failed: response is not a valid certificate package',
      detail: {
        title: 'Certificate detail',
        alias: 'Alias',
        cn: 'Common Name (CN)',
        san: 'SAN domains',
        notBefore: 'Valid from',
        notAfter: 'Expires',
        daysRemaining: 'Days remaining',
        expired: 'Expired',
        expectedRenewAt: 'Expected renewal',
        empty: 'No data'
      }
    }
  },
  request: {
    failed: 'Request failed',
    failedWithStatus: 'Request failed ({status})',
    expired: 'Session expired, please sign in again',
    network: 'Network error, please check the connection to the server',
    unknown: 'Unknown error'
  }
}
