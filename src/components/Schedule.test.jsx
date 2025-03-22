import React from 'react';
import { render, screen } from '@testing-library/react';
import { Schedule } from './Schedule';
import { Provider } from 'react-redux';
import store from '../store';

const renderWithProviders = (component) => {
  return render(
    <Provider store={store}>
      {component}
    </Provider>
  );
};

describe('Schedule component', () => {
  it('shows login prompt when not authenticated', () => {
    renderWithProviders(<Schedule />);
    expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
    expect(screen.getByText('Please login or signup to schedule appointments')).toBeInTheDocument();
  });
}); 