import React, { useContext } from 'react';
import { FormattedMessage } from 'react-intl';
import { useNavigate } from 'react-router-dom';
import formatted_message_rules from "@features/formatted_message_rules";
import * as styles from '@components/App.scss';
import { UserInfoContext } from '@features/user-info';

const AppMenu: React.FC = () => {
  const navigate = useNavigate();
  const userInfo = useContext(UserInfoContext);

  return (
    <nav className={styles.navigation_wrapper}>
      <ul>
        {userInfo && <li><span>{userInfo.name} ({userInfo.email})</span></li>}
        <li>
          <a onClick={() => navigate('/example', { relative: 'path' })}>
            <FormattedMessage
              id='example.menu.item'
              defaultMessage='Example'
              values={{
                ...formatted_message_rules,
              }}
            />
          </a>
        </li>
        <li>
          <a onClick={() => navigate('/credits', { relative: 'path' })}>
            <FormattedMessage
              id='credits.menu.item'
              defaultMessage='Example'
              values={{
                ...formatted_message_rules,
              }}
            />
          </a>
        </li>
        {userInfo && <li>
          <a href={userInfo.logout_url}>
            <FormattedMessage
              id='app.logout'
              defaultMessage='Logout'
              values={{
                ...formatted_message_rules,
              }}
            />
          </a>
        </li>}
      </ul>
    </nav>
  );
};

export default AppMenu;
