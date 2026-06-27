import * as Location from 'expo-location';
import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react';

type Coordinates = {
  lat: number;
  lng: number;
};

type LocationContextValue = {
  coordinates: Coordinates | null;
  permissionStatus: Location.PermissionStatus | null;
  backgroundPermissionStatus: Location.PermissionStatus | null;
  loading: boolean;
  dismissLocationPermission: () => void;
  requestCurrentLocation: (scope?: 'foreground' | 'background') => Promise<Coordinates | null>;
};

const LocationContext = createContext<LocationContextValue | null>(null);

export function LocationProvider({ children }: PropsWithChildren) {
  const [coordinates, setCoordinates] = useState<Coordinates | null>(null);
  const [permissionStatus, setPermissionStatus] = useState<Location.PermissionStatus | null>(null);
  const [backgroundPermissionStatus, setBackgroundPermissionStatus] =
    useState<Location.PermissionStatus | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    async function restorePermission() {
      const [foreground, background] = await Promise.all([
        Location.getForegroundPermissionsAsync(),
        Location.getBackgroundPermissionsAsync(),
      ]);
      setPermissionStatus(foreground.status);
      setBackgroundPermissionStatus(background.status);

      if (foreground.status === Location.PermissionStatus.GRANTED) {
        const lastKnown = await Location.getLastKnownPositionAsync();
        if (lastKnown) {
          setCoordinates({
            lat: lastKnown.coords.latitude,
            lng: lastKnown.coords.longitude,
          });
        }
      }
    }

    restorePermission();
  }, []);

  const dismissLocationPermission = useCallback(() => {
    setPermissionStatus(Location.PermissionStatus.DENIED);
    setCoordinates(null);
  }, []);

  const requestCurrentLocation = useCallback(async (scope: 'foreground' | 'background' = 'foreground') => {
    setLoading(true);

    try {
      let foreground = await Location.getForegroundPermissionsAsync();
      if (foreground.status !== Location.PermissionStatus.GRANTED) {
        foreground = await Location.requestForegroundPermissionsAsync();
      }
      setPermissionStatus(foreground.status);

      if (foreground.status !== Location.PermissionStatus.GRANTED) {
        setCoordinates(null);
        return null;
      }

      if (scope === 'background') {
        const background = await Location.requestBackgroundPermissionsAsync();
        setBackgroundPermissionStatus(background.status);
      }

      const position = await Location.getCurrentPositionAsync({
        accuracy: Location.Accuracy.Balanced,
      });

      const nextCoordinates = {
        lat: position.coords.latitude,
        lng: position.coords.longitude,
      };

      setCoordinates(nextCoordinates);
      return nextCoordinates;
    } finally {
      setLoading(false);
    }
  }, []);

  const value = useMemo(
    () => ({
      coordinates,
      permissionStatus,
      backgroundPermissionStatus,
      loading,
      dismissLocationPermission,
      requestCurrentLocation,
    }),
    [
      backgroundPermissionStatus,
      coordinates,
      dismissLocationPermission,
      loading,
      permissionStatus,
      requestCurrentLocation,
    ],
  );

  return <LocationContext.Provider value={value}>{children}</LocationContext.Provider>;
}

export function useCurrentLocation() {
  const value = useContext(LocationContext);

  if (!value) {
    throw new Error('useCurrentLocation must be used within LocationProvider');
  }

  return value;
}
