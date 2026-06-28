import { useEffect, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, TextInput, View } from 'react-native';
import { Link } from 'expo-router';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { useAuth } from '@/contexts/auth-context';
import { useCurrentLocation } from '@/contexts/location-context';

type AuthMode = 'intro' | 'login' | 'register' | 'otp';

export default function HomeScreen() {
  const {
    coordinates,
    dismissLocationPermission,
    loading,
    permissionStatus,
    requestCurrentLocation,
  } = useCurrentLocation();
  const {
    booting: authBooting,
    login,
    register,
    requestOTP,
    shouldAskLogin,
    skipLoginPrompt,
    user,
    verifyOTP,
  } = useAuth();
  const [showLocationPopup, setShowLocationPopup] = useState(false);
  const [showLoginPopup, setShowLoginPopup] = useState(false);
  const [authMode, setAuthMode] = useState<AuthMode>('intro');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [debugOtp, setDebugOtp] = useState<string | null>(null);
  const [authLoading, setAuthLoading] = useState(false);
  const [authError, setAuthError] = useState<string | null>(null);

  useEffect(() => {
    async function checkLoginPrompt() {
      if (!authBooting && !user && (await shouldAskLogin())) {
        setShowLoginPopup(true);
      }
    }

    checkLoginPrompt();
  }, [authBooting, shouldAskLogin, user]);

  useEffect(() => {
    if (!showLoginPopup && !coordinates && !permissionStatus) {
      const timer = setTimeout(() => setShowLocationPopup(true), 450);
      return () => clearTimeout(timer);
    }
  }, [coordinates, permissionStatus, showLoginPopup]);

  const askLocation = async (scope: 'foreground' | 'background') => {
    await requestCurrentLocation(scope);
    setShowLocationPopup(false);
  };

  const closeLoginForToday = async () => {
    await skipLoginPrompt();
    setShowLoginPopup(false);
  };

  const submitLogin = async () => {
    setAuthLoading(true);
    setAuthError(null);
    try {
      if (authMode === 'register') {
        await register(phone, password);
      } else if (authMode === 'otp') {
        await verifyOTP(phone, otp);
      } else {
        await login(phone, password);
      }
      setShowLoginPopup(false);
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Không đăng nhập được');
    } finally {
      setAuthLoading(false);
    }
  };

  const submitRequestOTP = async () => {
    setAuthLoading(true);
    setAuthError(null);
    try {
      const response = await requestOTP(phone);
      setDebugOtp(response.debug_otp ?? null);
      setAuthMode('otp');
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Không gửi được OTP');
    } finally {
      setAuthLoading(false);
    }
  };

  return (
    <ThemedView style={styles.container}>
      <Link href="/today-eat" asChild>
        <Pressable style={styles.card}>
          <ThemedText type="title" style={styles.cardTitle}>
            👀 Hôm nay ăn gì
          </ThemedText>
        </Pressable>
      </Link>

      {user ? (
        <Link href="/viewed-restaurants" asChild>
          <Pressable style={styles.card}>
            <ThemedText type="title" style={styles.cardTitle}>
              🍽️ Quán ăn bạn đã xem
            </ThemedText>
            <ThemedText style={styles.cardDesc}>Danh sách các quán đã mở hôm nay</ThemedText>
          </Pressable>
        </Link>
      ) : (
        <Pressable style={[styles.card, styles.disabledCard]}>
          <ThemedText type="title" style={styles.cardTitle}>
            🍽️ Quán ăn bạn đã xem
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Đăng nhập để xem mục này</ThemedText>
        </Pressable>
      )}

      {user ? (
        <Link href="/saved-restaurants" asChild>
          <Pressable style={styles.card}>
            <ThemedText type="title" style={styles.cardTitle}>
              💾 Quán ăn đã lưu
            </ThemedText>
            <ThemedText style={styles.cardDesc}>Những quán bạn muốn quay lại</ThemedText>
          </Pressable>
        </Link>
      ) : (
        <Pressable style={[styles.card, styles.disabledCard]}>
          <ThemedText type="title" style={styles.cardTitle}>
            💾 Quán ăn đã lưu
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Đăng nhập để lưu quán</ThemedText>
        </Pressable>
      )}

      <Modal transparent visible={showLocationPopup} animationType="fade">
        <View style={styles.overlay}>
          <View style={styles.popup}>
            <ThemedText type="title" style={styles.popupTitle}>
              Cho WWETD biết bạn đang ở đâu?
            </ThemedText>
            <ThemedText style={styles.popupDesc}>
              Mình sẽ dùng vị trí hiện tại để gợi ý quán ăn gần bạn hơn. Nếu bỏ qua, app vẫn chọn
              quán dựa trên IP.
            </ThemedText>

            <Pressable
              style={styles.allowBtn}
              onPress={() => askLocation('foreground')}
              disabled={loading}
            >
              <ThemedText style={styles.allowText}>
                {loading ? 'Đang lấy vị trí...' : 'Trong khi dùng ứng dụng'}
              </ThemedText>
            </Pressable>

            <Pressable
              style={[styles.allowBtn, styles.alwaysBtn]}
              onPress={() => askLocation('background')}
              disabled={loading}
            >
              <ThemedText style={styles.allowText}>Luôn luôn cho phép</ThemedText>
            </Pressable>

            <Pressable
              style={styles.skipBtn}
              onPress={() => {
                dismissLocationPermission();
                setShowLocationPopup(false);
              }}
            >
              <ThemedText style={styles.skipText}>Không cho phép</ThemedText>
            </Pressable>
          </View>
        </View>
      </Modal>

      <Modal transparent visible={showLoginPopup} animationType="fade">
        <View style={styles.overlay}>
          <View style={styles.popup}>
            {authMode === 'intro' ? (
              <>
                <Image source={require('@/assets/images/icon.png')} style={styles.appIcon} />

                <Pressable style={styles.allowBtn} onPress={() => setAuthMode('login')}>
                  <ThemedText style={styles.allowText}>Đăng nhập</ThemedText>
                </Pressable>

                <Pressable style={styles.skipBtn} onPress={closeLoginForToday}>
                  <ThemedText style={styles.skipText}>Để sau, nhắc lại sau 24 giờ</ThemedText>
                </Pressable>
              </>
            ) : (
              <>
                <ThemedText type="title" style={styles.popupTitle}>
                  {authMode === 'register'
                    ? 'Tạo tài khoản'
                    : authMode === 'otp'
                      ? 'Nhập OTP'
                      : 'Đăng nhập'}
                </ThemedText>

                <TextInput
                  keyboardType="phone-pad"
                  onChangeText={setPhone}
                  placeholder="Số điện thoại"
                  placeholderTextColor="#8a9678"
                  style={styles.input}
                  value={phone}
                />

                {authMode === 'otp' ? (
                  <>
                    <TextInput
                      keyboardType="number-pad"
                      maxLength={6}
                      onChangeText={setOtp}
                      placeholder="Mã OTP"
                      placeholderTextColor="#8a9678"
                      style={styles.input}
                      value={otp}
                    />
                    {debugOtp ? (
                      <ThemedText style={styles.debugOtp}>OTP dev: {debugOtp}</ThemedText>
                    ) : null}
                  </>
                ) : (
                  <TextInput
                    onChangeText={setPassword}
                    placeholder="Mật khẩu"
                    placeholderTextColor="#8a9678"
                    secureTextEntry
                    style={styles.input}
                    value={password}
                  />
                )}

                {authError ? <ThemedText style={styles.errorText}>{authError}</ThemedText> : null}

                <Pressable style={styles.allowBtn} onPress={submitLogin} disabled={authLoading}>
                  <ThemedText style={styles.allowText}>
                    {authLoading
                      ? 'Đang xử lý...'
                      : authMode === 'register'
                        ? 'Đăng ký'
                        : authMode === 'otp'
                          ? 'Xác thực OTP'
                          : 'Đăng nhập'}
                  </ThemedText>
                </Pressable>

                {authMode === 'login' ? (
                  <Pressable
                    style={styles.secondaryBtn}
                    onPress={submitRequestOTP}
                    disabled={authLoading}
                  >
                    <ThemedText style={styles.skipText}>Đăng nhập bằng OTP</ThemedText>
                  </Pressable>
                ) : null}

                <Pressable
                  style={styles.secondaryBtn}
                  onPress={() => {
                    setAuthError(null);
                    setDebugOtp(null);
                    setAuthMode(authMode === 'register' ? 'login' : 'register');
                  }}
                >
                  <ThemedText style={styles.skipText}>
                    {authMode === 'register' ? 'Tôi đã có tài khoản' : 'Tạo tài khoản mới'}
                  </ThemedText>
                </Pressable>

                <Pressable style={styles.skipBtn} onPress={closeLoginForToday}>
                  <ThemedText style={styles.skipText}>Để sau</ThemedText>
                </Pressable>
              </>
            )}
          </View>
        </View>
      </Modal>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  card: {
    width: '100%',
    maxWidth: 350,
    height: 150,
    backgroundColor: '#a7c068',
    borderRadius: 20,
    padding: 20,
    justifyContent: 'center',
    marginBottom: 20,

    shadowColor: '#000',
    shadowOpacity: 0.15,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 4 },
  },
  disabledCard: {
    opacity: 0.45,
  },
  cardTitle: {
    color: '#fff',
    marginBottom: 8,
  },
  cardDesc: {
    color: '#f4f4f4',
    fontSize: 14,
  },
  overlay: {
    alignItems: 'center',
    backgroundColor: 'rgba(19,31,12,0.48)',
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },
  popup: {
    backgroundColor: '#fbfff3',
    borderRadius: 22,
    padding: 20,
    width: '100%',
  },
  popupTitle: {
    color: '#21320f',
    marginBottom: 8,
  },
  appIcon: {
    alignSelf: 'center',
    borderRadius: 18,
    height: 72,
    marginBottom: 18,
    width: 72,
  },
  popupDesc: {
    color: '#667653',
    fontSize: 15,
    lineHeight: 22,
    marginBottom: 18,
  },
  allowBtn: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderRadius: 16,
    paddingVertical: 14,
  },
  allowText: {
    color: '#fffdf5',
    fontWeight: '800',
  },
  alwaysBtn: {
    backgroundColor: '#496a24',
    marginTop: 10,
  },
  skipBtn: {
    alignItems: 'center',
    paddingVertical: 13,
  },
  skipText: {
    color: '#496a24',
    fontWeight: '800',
  },
  secondaryBtn: {
    alignItems: 'center',
    paddingTop: 13,
  },
  input: {
    backgroundColor: '#eef6df',
    borderRadius: 14,
    color: '#21320f',
    fontSize: 16,
    marginBottom: 10,
    paddingHorizontal: 14,
    paddingVertical: 13,
  },
  errorText: {
    color: '#c2410c',
    fontSize: 13,
    fontWeight: '700',
    marginBottom: 10,
  },
  debugOtp: {
    color: '#667653',
    fontSize: 13,
    fontWeight: '700',
    marginBottom: 10,
  },
});
