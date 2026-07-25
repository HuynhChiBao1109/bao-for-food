import { Ionicons } from '@expo/vector-icons';
import * as ImagePicker from 'expo-image-picker';
import { useRouter } from 'expo-router';
import { useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { ThemedText } from '@/components/themed-text';
import { apiAssetURL } from '@/constants/api';
import { useAuth } from '@/contexts/auth-context';

const MAX_AVATAR_SIZE = 5 * 1024 * 1024;

function formatDate(value?: string) {
  if (!value) return 'Chưa có';

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Chưa có';
  return new Intl.DateTimeFormat('vi-VN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

export default function ProfileScreen() {
  const router = useRouter();
  const { refreshUser, updateName, uploadAvatar, user } = useAuth();
  const [name, setName] = useState(user?.name ?? '');
  const [loading, setLoading] = useState(true);
  const [savingName, setSavingName] = useState(false);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const avatarURL = useMemo(() => apiAssetURL(user?.avatar), [user?.avatar]);
  const avatarLabel =
    user?.name?.trim().slice(0, 2).toLocaleUpperCase('vi-VN') || user?.phone?.slice(-2) || 'BA';

  useEffect(() => {
    refreshUser()
      .catch((err) => {
        setMessage(err instanceof Error ? err.message : 'Không tải được thông tin cá nhân');
      })
      .finally(() => setLoading(false));
  }, [refreshUser]);

  useEffect(() => {
    setName(user?.name ?? '');
  }, [user?.name]);

  const saveName = async () => {
    const nextName = name.trim();
    if (!nextName) {
      setMessage('Tên không được để trống');
      return;
    }

    setSavingName(true);
    setMessage(null);
    try {
      await updateName(nextName);
      setMessage('Đã cập nhật tên');
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Không cập nhật được tên');
    } finally {
      setSavingName(false);
    }
  };

  const chooseAvatar = async () => {
    setMessage(null);
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      setMessage('BAO cần quyền truy cập thư viện ảnh để đổi avatar');
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      allowsEditing: true,
      aspect: [1, 1],
      mediaTypes: ['images'],
      quality: 0.85,
    });
    if (result.canceled) return;

    const asset = result.assets[0];
    if (asset.fileSize && asset.fileSize > MAX_AVATAR_SIZE) {
      setMessage('Ảnh đại diện phải nhỏ hơn hoặc bằng 5 MB');
      return;
    }

    setUploadingAvatar(true);
    try {
      await uploadAvatar({
        uri: asset.uri,
        fileName: asset.fileName,
        mimeType: asset.mimeType,
        file: asset.file,
      });
      setMessage('Đã cập nhật ảnh đại diện');
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Không cập nhật được ảnh đại diện');
    } finally {
      setUploadingAvatar(false);
    }
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.container}
      >
        <View style={styles.header}>
          <Pressable
            accessibilityLabel="Quay lại"
            hitSlop={10}
            onPress={() => router.back()}
            style={styles.iconButton}
          >
            <Ionicons color="#21320f" name="arrow-back" size={24} />
          </Pressable>
          <ThemedText style={styles.headerTitle}>Thông tin cá nhân</ThemedText>
          <View style={styles.iconButton} />
        </View>

        <ScrollView
          contentContainerStyle={styles.content}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          {loading ? (
            <ActivityIndicator color="#496a24" size="large" style={styles.loader} />
          ) : (
            <>
              <Pressable
                accessibilityLabel="Đổi ảnh đại diện"
                onPress={chooseAvatar}
                style={styles.avatarButton}
                disabled={uploadingAvatar}
              >
                {avatarURL ? (
                  <Image source={{ uri: avatarURL }} style={styles.avatarImage} />
                ) : (
                  <View style={styles.avatarFallback}>
                    <ThemedText style={styles.avatarText}>{avatarLabel}</ThemedText>
                  </View>
                )}
                <View style={styles.cameraButton}>
                  {uploadingAvatar ? (
                    <ActivityIndicator color="#fff" size="small" />
                  ) : (
                    <Ionicons color="#fff" name="camera" size={18} />
                  )}
                </View>
              </Pressable>

              <View style={styles.editSection}>
                <ThemedText style={styles.label}>Tên hiển thị</ThemedText>
                <TextInput
                  autoCapitalize="words"
                  maxLength={80}
                  onChangeText={setName}
                  onSubmitEditing={saveName}
                  placeholder="Tên của bạn"
                  placeholderTextColor="#8a9678"
                  returnKeyType="done"
                  style={styles.input}
                  value={name}
                />
                <Pressable
                  onPress={saveName}
                  disabled={savingName || !name.trim() || name.trim() === user?.name}
                  style={[
                    styles.saveButton,
                    (savingName || !name.trim() || name.trim() === user?.name) &&
                      styles.saveButtonDisabled,
                  ]}
                >
                  {savingName ? (
                    <ActivityIndicator color="#fff" size="small" />
                  ) : (
                    <Ionicons color="#fff" name="save-outline" size={18} />
                  )}
                  <ThemedText style={styles.saveButtonText}>Lưu thay đổi</ThemedText>
                </Pressable>
              </View>

              {message ? <ThemedText style={styles.message}>{message}</ThemedText> : null}

              <View style={styles.details}>
                <ProfileRow icon="call-outline" label="Số điện thoại" value={user?.phone} />
                <ProfileRow icon="finger-print-outline" label="ID người dùng" value={user?.id} />
                <ProfileRow
                  icon="calendar-outline"
                  label="Ngày tạo"
                  value={formatDate(user?.created_at)}
                />
                <ProfileRow
                  icon="time-outline"
                  label="Cập nhật gần nhất"
                  value={formatDate(user?.updated_at)}
                />
                <ProfileRow
                  icon="image-outline"
                  label="Đường dẫn avatar"
                  value={user?.avatar || 'Chưa có'}
                />
              </View>
            </>
          )}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

function ProfileRow({
  icon,
  label,
  value,
}: {
  icon: keyof typeof Ionicons.glyphMap;
  label: string;
  value?: string;
}) {
  return (
    <View style={styles.detailRow}>
      <Ionicons color="#496a24" name={icon} size={20} />
      <View style={styles.detailContent}>
        <ThemedText style={styles.detailLabel}>{label}</ThemedText>
        <ThemedText selectable style={styles.detailValue}>
          {value || 'Chưa có'}
        </ThemedText>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    backgroundColor: '#f7fbec',
    flex: 1,
  },
  container: {
    flex: 1,
  },
  header: {
    alignItems: 'center',
    borderBottomColor: '#dce8ca',
    borderBottomWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 56,
    paddingHorizontal: 16,
  },
  iconButton: {
    alignItems: 'center',
    height: 40,
    justifyContent: 'center',
    width: 40,
  },
  headerTitle: {
    color: '#21320f',
    fontSize: 18,
    fontWeight: '800',
  },
  content: {
    alignSelf: 'center',
    padding: 20,
    paddingBottom: 40,
    width: '100%',
    maxWidth: 560,
  },
  loader: {
    marginTop: 80,
  },
  avatarButton: {
    alignSelf: 'center',
    height: 112,
    marginBottom: 28,
    position: 'relative',
    width: 112,
  },
  avatarImage: {
    borderColor: '#dbeabf',
    borderRadius: 56,
    borderWidth: 4,
    height: 112,
    width: 112,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#496a24',
    borderColor: '#dbeabf',
    borderRadius: 56,
    borderWidth: 4,
    height: 112,
    justifyContent: 'center',
    width: 112,
  },
  avatarText: {
    color: '#fff',
    fontSize: 30,
    fontWeight: '900',
  },
  cameraButton: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderColor: '#f7fbec',
    borderRadius: 20,
    borderWidth: 3,
    bottom: -2,
    height: 40,
    justifyContent: 'center',
    position: 'absolute',
    right: -2,
    width: 40,
  },
  editSection: {
    borderBottomColor: '#dce8ca',
    borderBottomWidth: StyleSheet.hairlineWidth,
    paddingBottom: 24,
  },
  label: {
    color: '#496a24',
    fontSize: 13,
    fontWeight: '800',
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#eef6df',
    borderColor: '#d4e3bc',
    borderRadius: 8,
    borderWidth: 1,
    color: '#21320f',
    fontSize: 16,
    marginBottom: 12,
    paddingHorizontal: 14,
    paddingVertical: 13,
  },
  saveButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    backgroundColor: '#e67e45',
    borderRadius: 8,
    flexDirection: 'row',
    gap: 8,
    minHeight: 44,
    paddingHorizontal: 16,
  },
  saveButtonDisabled: {
    backgroundColor: '#b8c1ad',
  },
  saveButtonText: {
    color: '#fff',
    fontWeight: '800',
  },
  message: {
    color: '#496a24',
    fontSize: 14,
    fontWeight: '700',
    marginTop: 16,
  },
  details: {
    marginTop: 22,
  },
  detailRow: {
    alignItems: 'flex-start',
    borderBottomColor: '#e1ead4',
    borderBottomWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: 12,
    paddingVertical: 14,
  },
  detailContent: {
    flex: 1,
  },
  detailLabel: {
    color: '#667653',
    fontSize: 12,
    fontWeight: '700',
    marginBottom: 3,
  },
  detailValue: {
    color: '#21320f',
    fontSize: 15,
  },
});
