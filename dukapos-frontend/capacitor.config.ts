import { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'com.dukapos.app',
  appName: 'DukaPOS',
  webDir: 'dist',
  server: {
    androidScheme: 'https',
    url: process.env.CAPACITOR_SERVER_URL || 'http://localhost:3000',
    cleartext: true
  },
  plugins: {
    SplashScreen: {
      launchShowDuration: 3000,
      backgroundColor: '#00A650',
      showSpinner: true,
      spinnerColor: '#ffffff',
      splashFullScreen: true,
      splashImmersive: true
    },
    PushNotifications: {
      presentationOptions: ['badge', 'sound', 'alert']
    }
  },
  android: {
    buildToolsVersion: '34.0.0',
    minSdkVersion: 23,
    targetSdkVersion: 34,
    useAndroidX: true,
    allowMixedContent: true
  },
  ios: {
    minVersion: '13.0',
    devices: ['iphone']
  }
}

export default config
