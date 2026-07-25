import { useState } from 'react';
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

type NameFormModalProps = {
  visible: boolean;
  updateName: (name: string) => Promise<void>;
};

export function NameFormModal({ visible, updateName }: NameFormModalProps) {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setError('Bạn cần nhập tên để tiếp tục');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await updateName(trimmedName);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không lưu được tên');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal transparent visible={visible} animationType="fade" onRequestClose={() => undefined}>
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
            <Image source={require('@/assets/images/icon.png')} style={styles.appIcon} />
            <ThemedText type="title" style={styles.title}>
              BAO có thể gọi bạn bằng gì?
            </ThemedText>

            <TextInput
              autoCapitalize="words"
              autoFocus
              maxLength={80}
              onChangeText={setName}
              onSubmitEditing={submit}
              placeholder="Tên của bạn"
              placeholderTextColor="#8a9678"
              returnKeyType="done"
              style={styles.input}
              value={name}
            />

            {error ? <ThemedText style={styles.errorText}>{error}</ThemedText> : null}

            <Pressable style={styles.primaryButton} onPress={submit} disabled={loading}>
              <ThemedText style={styles.primaryText}>
                {loading ? 'Đang lưu...' : 'Tiếp tục'}
              </ThemedText>
            </Pressable>
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
    backgroundColor: 'rgba(19,31,12,0.58)',
    flexGrow: 1,
    justifyContent: 'center',
    padding: 24,
  },
  popup: {
    backgroundColor: '#fbfff3',
    borderRadius: 8,
    maxWidth: 480,
    padding: 20,
    width: '100%',
  },
  appIcon: {
    alignSelf: 'center',
    borderRadius: 16,
    height: 72,
    marginBottom: 14,
    width: 72,
  },
  title: {
    color: '#21320f',
    fontSize: 22,
    lineHeight: 30,
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
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderRadius: 8,
    paddingVertical: 14,
  },
  primaryText: {
    color: '#fffdf5',
    fontWeight: '800',
  },
  errorText: {
    color: '#c2410c',
    fontSize: 13,
    fontWeight: '700',
    marginBottom: 10,
  },
});
