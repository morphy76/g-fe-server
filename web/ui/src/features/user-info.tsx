import React, { PropsWithChildren, createContext } from 'react';
import { useQuery } from 'react-query';
import { auth_client } from '@features/axios';

export type UserInfo = {
  email:string;
  family_name:string;
  given_name:string;
  name:string;
  preferred_username:string;
  logout_url:string;
};

const fetchUserInfo = async (): Promise<UserInfo> => {
  const res = await auth_client.get('/info');
  return res.data;
};

export const useUserInfoQuery = () => {
  return useQuery<UserInfo, Error>("user_info", fetchUserInfo);
};

export const UserInfoContext = createContext<UserInfo | undefined>(undefined);

type UserInfoContextProviderProps = PropsWithChildren<{
  userInfo: UserInfo | undefined
}>;
export const UserInfoContextProvider: React.FC<UserInfoContextProviderProps> = ({ userInfo, children }) => (
  <UserInfoContext.Provider value={userInfo}>
    {children}
  </UserInfoContext.Provider>
);
