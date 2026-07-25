import { Ionicons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, View } from 'react-native';
import { Link, useRouter } from 'expo-router';

import { AuthFormModal } from '@/components/auth-form-modal';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { apiAssetURL } from '@/constants/api';
import { useAuth } from '@/contexts/auth-context';
import { useCurrentLocation } from '@/contexts/location-context';

export default function HomeScreen() {
  const router = useRouter();
  const {
    coordinates,
    dismissLocationPermission,
    loading,
    permissionStatus,
    requestCurrentLocation,
  } = useCurrentLocation();
  const {
    booting: authBooting,
    clearAuth,
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
  const [showUserMenu, setShowUserMenu] = useState(false);
  const avatarURL = apiAssetURL(user?.avatar);
  const avatarLabel =
    user?.name?.trim().slice(0, 2).toLocaleUpperCase('vi-VN') || user?.phone?.slice(-2) || 'BA';

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

  const openLoginForm = () => {
    setShowLoginPopup(true);
  };

  const logout = async () => {
    setShowUserMenu(false);
    await skipLoginPrompt();
    await clearAuth();
  };

  return (
    <ThemedView style={styles.container}>
      {showUserMenu ? (
        <Pressable
          accessibilityLabel="Đóng menu người dùng"
          onPress={() => setShowUserMenu(false)}
          style={styles.menuBackdrop}
        />
      ) : null}

      <View style={styles.header}>
        <View style={styles.brand}>
          <Image source={require('@/assets/images/icon.png')} style={styles.headerIcon} />
          <View>
            <ThemedText style={styles.brandName}>BAO</ThemedText>
            <ThemedText style={styles.brandSub}>Hôm nay ăn gì?</ThemedText>
          </View>
        </View>

        {user ? (
          <>
            <Pressable
              accessibilityLabel="Mở menu người dùng"
              onPress={() => setShowUserMenu((current) => !current)}
              style={styles.avatar}
            >
              {avatarURL ? (
                <Image source={{ uri: avatarURL }} style={styles.avatarImage} />
              ) : (
                <ThemedText style={styles.avatarText}>{avatarLabel}</ThemedText>
              )}
            </Pressable>

            {showUserMenu ? (
              <View style={styles.userMenu}>
                <View style={styles.userSummary}>
                  <ThemedText numberOfLines={1} style={styles.userName}>
                    {user.name}
                  </ThemedText>
                  <ThemedText style={styles.userPhone}>{user.phone}</ThemedText>
                </View>

                <Pressable
                  onPress={() => {
                    setShowUserMenu(false);
                    router.push('/profile');
                  }}
                  style={styles.menuItem}
                >
                  <Ionicons color="#496a24" name="person-outline" size={20} />
                  <ThemedText style={styles.menuItemText}>Thông tin cá nhân</ThemedText>
                </Pressable>

                <Pressable onPress={logout} style={[styles.menuItem, styles.logoutItem]}>
                  <Ionicons color="#c2410c" name="log-out-outline" size={20} />
                  <ThemedText style={styles.logoutText}>Đăng xuất</ThemedText>
                </Pressable>
              </View>
            ) : null}
          </>
        ) : (
          <Pressable style={styles.loginBtn} onPress={openLoginForm}>
            <ThemedText style={styles.loginBtnText}>Đăng nhập</ThemedText>
          </Pressable>
        )}
      </View>

      <View style={styles.content}>
        <Link href="/today-eat" asChild>
          <Pressable style={styles.card}>
            <ThemedText type="title" style={styles.cardTitle}>
              👀 Hôm nay ăn gì
            </ThemedText>
          </Pressable>
        </Link>

        {/* {user ? (
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
      )} */}
      </View>

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

      <AuthFormModal
        visible={showLoginPopup}
        onClose={closeLoginForToday}
        onSuccess={() => setShowLoginPopup(false)}
        login={login}
        register={register}
        requestOTP={requestOTP}
        verifyOTP={verifyOTP}
      />
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f7fbec',
    padding: 20,
  },
  header: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingTop: 42,
    width: '100%',
    zIndex: 20,
  },
  brand: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 12,
  },
  headerIcon: {
    borderRadius: 16,
    height: 52,
    width: 52,
  },
  brandName: {
    color: '#21320f',
    fontSize: 22,
    fontWeight: '900',
  },
  brandSub: {
    color: '#667653',
    fontSize: 13,
    fontWeight: '700',
    marginTop: 2,
  },
  loginBtn: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderRadius: 18,
    minWidth: 104,
    paddingHorizontal: 14,
    paddingVertical: 11,
  },
  loginBtnText: {
    color: '#fffdf5',
    fontWeight: '900',
  },
  avatar: {
    alignItems: 'center',
    backgroundColor: '#496a24',
    borderColor: '#dbeabf',
    borderRadius: 24,
    borderWidth: 3,
    height: 48,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 48,
  },
  avatarImage: {
    height: '100%',
    width: '100%',
  },
  avatarText: {
    color: '#fffdf5',
    fontSize: 15,
    fontWeight: '900',
  },
  menuBackdrop: {
    bottom: 0,
    left: 0,
    position: 'absolute',
    right: 0,
    top: 0,
    zIndex: 10,
  },
  userMenu: {
    backgroundColor: '#fbfff3',
    borderColor: '#d9e6c5',
    borderRadius: 8,
    borderWidth: 1,
    position: 'absolute',
    right: 0,
    shadowColor: '#17200f',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.18,
    shadowRadius: 12,
    top: 98,
    width: 240,
    zIndex: 30,
    elevation: 8,
  },
  userSummary: {
    borderBottomColor: '#e1ead4',
    borderBottomWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: 14,
    paddingVertical: 12,
  },
  userName: {
    color: '#21320f',
    fontSize: 15,
    fontWeight: '900',
  },
  userPhone: {
    color: '#667653',
    fontSize: 13,
    marginTop: 2,
  },
  menuItem: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 10,
    minHeight: 48,
    paddingHorizontal: 14,
  },
  menuItemText: {
    color: '#21320f',
    fontWeight: '700',
  },
  logoutItem: {
    borderTopColor: '#e1ead4',
    borderTopWidth: StyleSheet.hairlineWidth,
  },
  logoutText: {
    color: '#c2410c',
    fontWeight: '800',
  },
  content: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    width: '100%',
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
});
