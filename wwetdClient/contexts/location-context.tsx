import * as Location from 'expo-location';
import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from 'react';

type Coordinates = {
  lat: number;
  lng: number;
};

type LocationContextValue = {
  coordinates: Coordinates | null;
  permissionStatus: Location.PermissionStatus | null;
  loading: boolean;
  requestCurrentLocation: () => Promise<Coordinates | null>;
};

const LocationContext = createContext<LocationContextValue | null>(null);

export function LocationProvider({ children }: PropsWithChildren) {
  const [coordinates, setCoordinates] = useState<Coordinates | null>(null);
  const [permissionStatus, setPermissionStatus] = useState<Location.PermissionStatus | null>(null);
  const [loading, setLoading] = useState(false);

  const requestCurrentLocation = useCallback(async () => {
    setLoading(true);

    try {
      const permission = await Location.requestForegroundPermissionsAsync();
      setPermissionStatus(permission.status);

      if (permission.status !== Location.PermissionStatus.GRANTED) {
        setCoordinates(null);
        return null;
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
      loading,
      requestCurrentLocation,
    }),
    [coordinates, loading, permissionStatus, requestCurrentLocation],
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
