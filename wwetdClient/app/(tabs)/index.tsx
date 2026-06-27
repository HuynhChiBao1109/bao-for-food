import { StyleSheet, Pressable } from 'react-native';
import { Link } from 'expo-router';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';

export default function HomeScreen() {
  return (
    <ThemedView style={styles.container}>
      <Link href="/rate-food" asChild>
        <Pressable style={styles.card}>
          <ThemedText type="title" style={styles.cardTitle}>
            🍽️ Đánh giá món ăn
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Chia sẻ cảm nhận về món bạn vừa ăn</ThemedText>
        </Pressable>
      </Link>

      <Link href="/today-eat" asChild>
        <Pressable style={styles.card}>
          <ThemedText type="title" style={styles.cardTitle}>
            👀 Hôm nay ăn gì
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Anh ơi bữa nay ăn gì</ThemedText>
        </Pressable>
      </Link>
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
  cardTitle: {
    color: '#fff',
    marginBottom: 8,
  },
  cardDesc: {
    color: '#f4f4f4',
    fontSize: 14,
  },
});
