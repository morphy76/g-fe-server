module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  moduleNameMapper: {
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$':
      '<rootDir>/src/__mocks__/fileMock.ts',
    '\\.(css|less|scss)$': '<rootDir>/src/__mocks__/styleMock.ts',
    '\\.bundle.json$': '<rootDir>/src/__mocks__/dataMock.ts',
    "^@features/(.*)$": "<rootDir>/src/features/$1",
    "^@components/(.*)$": "<rootDir>/src/components/$1"
  },
  collectCoverageFrom: ['**/*.{ts,tsx}', '!**/*.d.ts', '!**/node_modules/**', '!**/vendor/**', '!src/index.tsx'],
  coverageDirectory: '../../../target/coverage',
  coverageThreshold: {
    global: {
      branches: 75,
      functions: 75,
      lines: 75,
      statements: 75,
    },
    './src/features/react-logger.tsx': {
      branches: 0,
    },
    './src/features/formatted_message_rules.tsx': {
      branches: 0,
      functions: 0,
      lines: 0,
      statements: 0,
    },
  },
};
