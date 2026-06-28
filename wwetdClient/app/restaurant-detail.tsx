import { API_BASE_URL } from '@/constants/api';
import { Fonts } from '@/constants/theme';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { useAuth } from '@/contexts/auth-context';
import { useCurrentLocation } from '@/contexts/location-context';
import { useLocalSearchParams } from 'expo-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  Linking,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';

const BACKGROUND = '#6f8f46';
const SURFACE = '#fbfff3';
const SURFACE_SOFT = '#eef6df';
const INK = '#21320f';
const MUTED = '#667653';
const PRIMARY = '#496a24';

const FALLBACK_IMAGE = 'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe';

type Media = { url?: string; image_url?: string };
type PhotoGroup = { media?: Media[] };
type RestaurantDetail = {
  data_id?: string;
  title?: string;
  type?: string;
  rating?: number;
  reviews?: number | unknown[];
  location?: { address?: { full?: string } };
  link_google_maps?: string;
  photos?: PhotoGroup[];
  open_state?: { text?: string };
};

type DetailResponse = {
  data: {
    is_saved?: boolean;
    restaurant: RestaurantDetail;
  };
};

function restaurantImages(restaurant?: RestaurantDetail) {
  const urls =
    restaurant?.photos
      ?.flatMap((group) => group.media ?? [])
      .map((photo) => photo.url ?? photo.image_url)
      .filter((url): url is string => Boolean(url)) ?? [];

  return urls.length > 0 ? urls.slice(0, 8) : [FALLBACK_IMAGE];
}

function reviewCount(restaurant?: RestaurantDetail) {
  if (typeof restaurant?.reviews === 'number') return restaurant.reviews;
  return restaurant?.reviews?.length ?? 0;
}

export default function RestaurantDetailScreen() {
  const params = useLocalSearchParams<{ data_id?: string }>();
  const dataID = Array.isArray(params.data_id) ? params.data_id[0] : params.data_id;
  const { token, user } = useAuth();
  const { coordinates } = useCurrentLocation();
  const [restaurant, setRestaurant] = useState<RestaurantDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isSaved, setIsSaved] = useState(false);
  const [saveState, setSaveState] = useState<'idle' | 'loading' | 'success'>('idle');

  const images = useMemo(() => restaurantImages(restaurant ?? undefined), [restaurant]);

  const detailURL = useMemo(() => {
    if (!dataID) return null;
    const query = new URLSearchParams();
    if (coordinates) {
      query.set('lat', String(coordinates.lat));
      query.set('lng', String(coordinates.lng));
    }
    const suffix = query.toString();
    return `${API_BASE_URL}/api/v1/restaurants/${encodeURIComponent(dataID)}${suffix ? `?${suffix}` : ''}`;
  }, [coordinates, dataID]);

  const actionURL = useCallback(
    (action: 'saved' | 'viewed') => {
      if (!dataID) return null;
      const query = new URLSearchParams();
      if (coordinates) {
        query.set('lat', String(coordinates.lat));
        query.set('lng', String(coordinates.lng));
      }
      const suffix = query.toString();
      return `${API_BASE_URL}/api/v1/restaurants/${encodeURIComponent(dataID)}/${action}${
        suffix ? `?${suffix}` : ''
      }`;
    },
    [coordinates, dataID],
  );

  const recordAction = useCallback(
    async (action: 'saved' | 'viewed') => {
      if (!token) return false;
      const url = actionURL(action);
      if (!url) return false;
      const response = await fetch(url, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      return response.ok;
    },
    [actionURL, token],
  );

  useEffect(() => {
    async function load() {
      if (!detailURL) return;

      setLoading(true);
      setError(null);
      try {
        const response = await fetch(detailURL, {
          headers: token ? { Authorization: `Bearer ${token}` } : undefined,
        });
        if (!response.ok) {
          throw new Error('Không tải được chi tiết quán');
        }

        const payload = (await response.json()) as DetailResponse;
        setRestaurant(payload.data.restaurant);
        setIsSaved(Boolean(payload.data.is_saved));
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Không tải được chi tiết quán');
      } finally {
        setLoading(false);
      }
    }

    load();
  }, [detailURL, token]);

  const saveRestaurant = useCallback(async () => {
    if (!user || isSaved || saveState === 'loading') return;

    setSaveState('loading');
    const ok = await recordAction('saved');
    if (ok) {
      setIsSaved(true);
      setSaveState('success');
      setTimeout(() => setSaveState('idle'), 1500);
    } else {
      setSaveState('idle');
    }
  }, [isSaved, recordAction, saveState, user]);

  const openMap = useCallback(async () => {
    if (!restaurant?.link_google_maps) return;
    await recordAction('viewed');
    Linking.openURL(restaurant.link_google_maps);
  }, [recordAction, restaurant?.link_google_maps]);

  return (
    <ThemedView style={styles.container}>
      {loading ? (
        <View style={styles.center}>
          <ActivityIndicator color="#f8ffe9" />
          <ThemedText style={styles.loadingText}>Đang tải quán ăn...</ThemedText>
        </View>
      ) : error ? (
        <View style={styles.center}>
          <ThemedText type="title" style={styles.errorTitle}>
            Chưa mở được quán
          </ThemedText>
          <ThemedText style={styles.errorText}>{error}</ThemedText>
        </View>
      ) : (
        <>
          <ScrollView
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
          >
            <View style={styles.hero}>
              <Image source={{ uri: images[0] }} style={styles.heroImage} />
              <View style={styles.heroShade} />
              <View style={styles.heroContent}>
                {restaurant?.open_state?.text ? (
                  <ThemedText style={styles.kicker}>{restaurant.open_state.text}</ThemedText>
                ) : null}
                <ThemedText type="title" style={styles.title}>
                  {restaurant?.title ?? 'Quán ăn'}
                </ThemedText>
              </View>
            </View>

            <View style={styles.sheet}>
              <View style={styles.statsRow}>
                <View style={styles.statBox}>
                  <ThemedText style={styles.statValue}>
                    {restaurant?.rating ? restaurant.rating.toFixed(1) : '-'}
                  </ThemedText>
                  <ThemedText style={styles.statLabel}>Điểm</ThemedText>
                </View>
                <View style={styles.statBox}>
                  <ThemedText style={styles.statValue}>
                    {reviewCount(restaurant ?? undefined)}
                  </ThemedText>
                  <ThemedText style={styles.statLabel}>Đánh giá</ThemedText>
                </View>
              </View>

              <View style={styles.sectionBlock}>
                <ThemedText style={styles.sectionTitle}>Địa chỉ</ThemedText>
                <ThemedText style={styles.bodyText}>
                  {restaurant?.location?.address?.full ?? 'Chưa có địa chỉ'}
                </ThemedText>
              </View>

              {images.length > 1 ? (
                <View style={styles.sectionBlock}>
                  <ThemedText style={styles.sectionTitle}>Hình ảnh</ThemedText>
                  <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                    {images.map((image, index) => (
                      <Image
                        key={`${image}-${index}`}
                        source={{ uri: image }}
                        style={styles.galleryImage}
                      />
                    ))}
                  </ScrollView>
                </View>
              ) : null}
            </View>
          </ScrollView>

          <View style={styles.footer}>
            <Pressable style={[styles.actionBtn, styles.mapBtn]} onPress={openMap}>
              <ThemedText style={styles.mapBtnText}>Mở bản đồ</ThemedText>
            </Pressable>
            <Pressable
              disabled={!user || isSaved || saveState === 'loading'}
              style={[
                styles.actionBtn,
                user ? styles.saveBtn : styles.disabledBtn,
                isSaved ? styles.savedBtn : null,
              ]}
              onPress={saveRestaurant}
            >
              {saveState === 'loading' ? (
                <ActivityIndicator color={INK} size="small" />
              ) : (
                <ThemedText style={styles.saveBtnText}>
                  {saveState === 'success' ? 'Lưu thành công' : isSaved ? 'Đã lưu' : 'Lưu quán'}
                </ThemedText>
              )}
            </Pressable>
          </View>
        </>
      )}
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { backgroundColor: BACKGROUND, flex: 1 },
  center: { alignItems: 'center', flex: 1, justifyContent: 'center', padding: 24 },
  loadingText: { color: '#f8ffe9', fontFamily: Fonts.rounded, fontWeight: '800', marginTop: 12 },
  errorTitle: { color: '#fffdf5', fontFamily: Fonts.rounded, fontWeight: '900', marginBottom: 8 },
  errorText: { color: '#f5ffd9', textAlign: 'center' },
  scrollContent: { paddingBottom: 96 },
  hero: { height: 330, position: 'relative' },
  heroImage: { height: '100%', width: '100%' },
  heroShade: { ...StyleSheet.absoluteFillObject, backgroundColor: 'rgba(20,32,10,0.36)' },
  heroContent: { bottom: 26, left: 18, position: 'absolute', right: 18 },
  kicker: {
    color: '#f4ffd8',
    fontFamily: Fonts.rounded,
    fontSize: 14,
    fontWeight: '800',
    marginBottom: 8,
    textTransform: 'uppercase',
  },
  title: {
    color: '#fffdf5',
    fontFamily: Fonts.rounded,
    fontSize: 34,
    fontWeight: '900',
    lineHeight: 40,
  },
  sheet: {
    backgroundColor: SURFACE,
    borderTopLeftRadius: 26,
    borderTopRightRadius: 26,
    marginTop: -22,
    padding: 18,
  },
  statsRow: { flexDirection: 'row', gap: 10, marginBottom: 20 },
  statBox: {
    backgroundColor: SURFACE_SOFT,
    borderRadius: 14,
    flex: 1,
    justifyContent: 'center',
    minHeight: 78,
    padding: 10,
  },
  statValue: {
    color: PRIMARY,
    fontFamily: Fonts.rounded,
    fontSize: 21,
    fontWeight: '900',
    textAlign: 'center',
  },
  statLabel: {
    color: MUTED,
    fontFamily: Fonts.sans,
    fontSize: 12,
    fontWeight: '700',
    marginTop: 4,
    textAlign: 'center',
  },
  sectionBlock: { marginBottom: 20 },
  sectionTitle: {
    color: INK,
    fontFamily: Fonts.rounded,
    fontSize: 18,
    fontWeight: '900',
    marginBottom: 8,
  },
  bodyText: { color: MUTED, fontFamily: Fonts.sans, fontSize: 15, lineHeight: 22 },
  galleryImage: { borderRadius: 14, height: 210, marginRight: 10, width: 294 },
  footer: {
    backgroundColor: 'rgba(251,255,243,0.96)',
    bottom: 0,
    flexDirection: 'row',
    gap: 10,
    left: 0,
    padding: 14,
    position: 'absolute',
    right: 0,
  },
  actionBtn: { alignItems: 'center', borderRadius: 16, flex: 1, paddingVertical: 14 },
  mapBtn: { backgroundColor: SURFACE_SOFT },
  mapBtnText: { color: PRIMARY, fontFamily: Fonts.rounded, fontWeight: '900' },
  saveBtn: { backgroundColor: '#f4d35e' },
  savedBtn: { backgroundColor: '#dbeabf' },
  disabledBtn: { backgroundColor: '#d6dfc5', opacity: 0.55 },
  saveBtnText: { color: INK, fontFamily: Fonts.rounded, fontWeight: '900' },
});
