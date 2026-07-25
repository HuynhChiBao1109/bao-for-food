import { useEffect, useState } from 'react';
import {
  Image,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  TextInput,
  View,
} from 'react-native';

import { ThemedText } from '@/components/themed-text';

type AuthMode = 'login' | 'register' | 'otp';

type OTPResult = {
  debug_otp?: string;
};

type AuthFormModalProps = {
  visible: boolean;
  onClose: () => void | Promise<void>;
  onSuccess?: () => void | Promise<void>;
  login: (phone: string, password: string) => Promise<void>;
  register: (phone: string, password: string) => Promise<void>;
  requestOTP: (phone: string) => Promise<OTPResult>;
  verifyOTP: (phone: string, otp: string) => Promise<void>;
};

export function AuthFormModal({
  visible,
  onClose,
  onSuccess,
  login,
  register,
  requestOTP,
  verifyOTP,
}: AuthFormModalProps) {
  const [mode, setMode] = useState<AuthMode>('login');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [debugOtp, setDebugOtp] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!visible) return;

    setMode('login');
    setPassword('');
    setConfirmPassword('');
    setOtp('');
    setDebugOtp(null);
    setError(null);
  }, [visible]);

  const switchMode = (nextMode: AuthMode) => {
    setMode(nextMode);
    setError(null);
    setConfirmPassword('');
    if (nextMode !== 'otp') {
      setOtp('');
      setDebugOtp(null);
    }
  };

  const submit = async () => {
    if (!phone.trim()) {
      setError('Vui lòng nhập số điện thoại');
      return;
    }
    if (mode !== 'otp' && Array.from(password).length < 6) {
      setError('Mật khẩu phải có ít nhất 6 ký tự');
      return;
    }
    if (mode === 'register' && password !== confirmPassword) {
      setError('Mật khẩu xác nhận chưa trùng khớp');
      return;
    }
    if (mode === 'otp' && !otp.trim()) {
      setError('Vui lòng nhập mã OTP');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      if (mode === 'register') {
        await register(phone.trim(), password);
      } else if (mode === 'otp') {
        await verifyOTP(phone.trim(), otp.trim());
      } else {
        await login(phone.trim(), password);
      }
      await (onSuccess ?? onClose)();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không thể xác thực tài khoản');
    } finally {
      setLoading(false);
    }
  };

  const sendOTP = async () => {
    if (!phone.trim()) {
      setError('Nhập số điện thoại để nhận OTP');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await requestOTP(phone.trim());
      setDebugOtp(response.debug_otp ?? null);
      setMode('otp');
      setOtp('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không gửi được OTP');
    } finally {
      setLoading(false);
    }
  };

  const normalizedPhone = phone.replace(/\D/g, '');
  const phoneValid = normalizedPhone.length >= 8 && normalizedPhone.length <= 20;
  const passwordValid = Array.from(password).length >= 6;
  const passwordsMatch =
    passwordValid && confirmPassword.length > 0 && password === confirmPassword;
  const otpValid = /^\d{6}$/.test(otp);
  const canSubmit =
    !loading &&
    phoneValid &&
    (mode === 'otp' ? otpValid : passwordValid && (mode !== 'register' || passwordsMatch));

  return (
    <Modal transparent visible={visible} animationType="fade" onRequestClose={() => void onClose()}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.keyboardAvoiding}
      >
        <ScrollView
          bounces={false}
          contentContainerStyle={styles.overlay}
          keyboardShouldPersistTaps="handled"
        >
          <View style={styles.popup}>
            <Pressable
              accessibilityLabel="Đóng"
              hitSlop={10}
              onPress={() => void onClose()}
              style={styles.closeButton}
            >
              <ThemedText style={styles.closeText}>×</ThemedText>
            </Pressable>

            <Image source={require('@/assets/images/icon.png')} style={styles.appIcon} />
            <ThemedText type="title" style={styles.title}>
              {mode === 'register'
                ? 'Đăng ký'
                : mode === 'otp'
                  ? 'Đăng nhập bằng OTP'
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

            {mode === 'otp' ? (
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
            ) : mode === 'register' ? (
              <>
                <View style={styles.inputRow}>
                  <TextInput
                    onChangeText={setPassword}
                    placeholder="Mật khẩu"
                    placeholderTextColor="#8a9678"
                    secureTextEntry
                    style={styles.inputInRow}
                    value={password}
                  />
                  {passwordValid ? <ThemedText style={styles.checkIcon}>✓</ThemedText> : null}
                </View>
                <View style={styles.inputRow}>
                  <TextInput
                    onChangeText={setConfirmPassword}
                    onSubmitEditing={submit}
                    placeholder="Xác nhận mật khẩu"
                    placeholderTextColor="#8a9678"
                    secureTextEntry
                    style={styles.inputInRow}
                    value={confirmPassword}
                  />
                  {passwordsMatch ? <ThemedText style={styles.checkIcon}>✓</ThemedText> : null}
                </View>
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

            {error ? <ThemedText style={styles.errorText}>{error}</ThemedText> : null}

            <Pressable
              style={[styles.primaryButton, !canSubmit && styles.primaryButtonDisabled]}
              onPress={submit}
              disabled={!canSubmit}
            >
              <ThemedText style={styles.primaryText}>
                {loading
                  ? 'Đang xử lý...'
                  : mode === 'register'
                    ? 'Đăng ký'
                    : mode === 'otp'
                      ? 'Xác thực OTP'
                      : 'Đăng nhập'}
              </ThemedText>
            </Pressable>

            {mode === 'otp' ? (
              <Pressable style={styles.secondaryButton} onPress={sendOTP} disabled={loading}>
                <ThemedText style={styles.secondaryText}>Gửi lại OTP</ThemedText>
              </Pressable>
            ) : (
              <Pressable style={styles.secondaryButton} onPress={sendOTP} disabled={loading}>
                <ThemedText style={styles.secondaryText}>Đăng nhập bằng OTP</ThemedText>
              </Pressable>
            )}

            <Pressable
              style={styles.secondaryButton}
              onPress={() => switchMode(mode === 'register' ? 'login' : 'register')}
              disabled={loading}
            >
              <ThemedText style={styles.secondaryText}>
                {mode === 'register' ? 'Tôi đã có tài khoản' : 'Đăng ký tài khoản'}
              </ThemedText>
            </Pressable>

            {mode === 'otp' ? (
              <Pressable
                style={styles.secondaryButton}
                onPress={() => switchMode('login')}
                disabled={loading}
              >
                <ThemedText style={styles.secondaryText}>Đăng nhập bằng mật khẩu</ThemedText>
              </Pressable>
            ) : null}
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  keyboardAvoiding: {
    flex: 1,
  },
  overlay: {
    alignItems: 'center',
    backgroundColor: 'rgba(19,31,12,0.48)',
    flexGrow: 1,
    justifyContent: 'center',
    padding: 24,
  },
  popup: {
    backgroundColor: '#fbfff3',
    borderRadius: 8,
    padding: 20,
    position: 'relative',
    width: '100%',
    maxWidth: 480,
  },
  closeButton: {
    alignItems: 'center',
    height: 40,
    justifyContent: 'center',
    position: 'absolute',
    right: 8,
    top: 8,
    width: 40,
    zIndex: 1,
  },
  closeText: {
    color: '#496a24',
    fontSize: 28,
    fontWeight: '700',
    lineHeight: 30,
  },
  appIcon: {
    alignSelf: 'center',
    borderRadius: 16,
    height: 72,
    marginBottom: 12,
    width: 72,
  },
  title: {
    color: '#21320f',
    fontSize: 22,
    marginBottom: 18,
    textAlign: 'center',
  },
  input: {
    backgroundColor: '#eef6df',
    borderRadius: 8,
    color: '#21320f',
    fontSize: 16,
    marginBottom: 10,
    paddingHorizontal: 14,
    paddingVertical: 13,
  },
  inputRow: {
    alignItems: 'center',
    backgroundColor: '#eef6df',
    borderRadius: 8,
    flexDirection: 'row',
    marginBottom: 10,
    paddingRight: 14,
  },
  inputInRow: {
    color: '#21320f',
    flex: 1,
    fontSize: 16,
    paddingHorizontal: 14,
    paddingVertical: 13,
  },
  checkIcon: {
    color: '#3f7d20',
    fontSize: 20,
    fontWeight: '900',
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderRadius: 8,
    marginTop: 2,
    paddingVertical: 14,
  },
  primaryButtonDisabled: {
    backgroundColor: '#bdc5b2',
    opacity: 0.75,
  },
  primaryText: {
    color: '#fffdf5',
    fontWeight: '800',
  },
  secondaryButton: {
    alignItems: 'center',
    paddingTop: 14,
  },
  secondaryText: {
    color: '#496a24',
    fontWeight: '800',
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
