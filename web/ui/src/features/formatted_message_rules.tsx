import React, { ReactNode } from 'react';

const formatted_message_rules = {
  enlarge: (word: ReactNode) => <strong>{word}</strong>,
  i: (word: ReactNode) => <i>{word}</i>
};

export default formatted_message_rules;
