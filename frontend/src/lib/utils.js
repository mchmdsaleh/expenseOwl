export const colorPalette = [
  '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4',
  '#FFBE0B', '#FF006E', '#8338EC', '#3A86FF',
  '#FB5607', '#38B000', '#9B5DE5', '#F15BB5'
];

export const currencyBehaviors = {
  usd: { symbol: '$', useComma: false, useDecimals: true },
  eur: { symbol: '€', useComma: true, useDecimals: true },
  gbp: { symbol: '£', useComma: false, useDecimals: true },
  jpy: { symbol: '¥', useComma: false, useDecimals: false },
  cny: { symbol: '¥', useComma: false, useDecimals: true },
  krw: { symbol: '₩', useComma: false, useDecimals: false },
  inr: { symbol: '₹', useComma: false, useDecimals: true },
  rub: { symbol: '₽', useComma: true, useDecimals: true },
  brl: { symbol: 'R$', useComma: true, useDecimals: true },
  zar: { symbol: 'R', useComma: false, useDecimals: true },
  aed: { symbol: 'AED', useComma: false, useDecimals: true },
  aud: { symbol: 'A$', useComma: false, useDecimals: true },
  cad: { symbol: 'C$', useComma: false, useDecimals: true },
  chf: { symbol: 'Fr', useComma: false, useDecimals: true },
  hkd: { symbol: 'HK$', useComma: false, useDecimals: true },
  bdt: { symbol: '৳', useComma: false, useDecimals: true },
  sgd: { symbol: 'S$', useComma: false, useDecimals: true },
  thb: { symbol: '฿', useComma: false, useDecimals: true },
  try: { symbol: '₺', useComma: true, useDecimals: true },
  mxn: { symbol: 'Mex$', useComma: false, useDecimals: true },
  php: { symbol: '₱', useComma: false, useDecimals: true },
  pln: { symbol: 'zł', useComma: true, useDecimals: true },
  sek: { symbol: 'kr', useComma: false, useDecimals: true },
  nzd: { symbol: 'NZ$', useComma: false, useDecimals: true },
  dkk: { symbol: 'kr.', useComma: true, useDecimals: true },
  idr: { symbol: 'Rp', useComma: false, useDecimals: true },
  ils: { symbol: '₪', useComma: false, useDecimals: true },
  vnd: { symbol: '₫', useComma: true, useDecimals: false },
  myr: { symbol: 'RM', useComma: false, useDecimals: true }
};

export function formatCurrency(amount, currency = 'usd') {
  const behavior = currencyBehaviors[currency] || { symbol: '$', useComma: false, useDecimals: true };
  const isNegative = amount < 0;
  const absAmount = Math.abs(amount);
  const options = {
    minimumFractionDigits: behavior.useDecimals ? 2 : 0,
    maximumFractionDigits: behavior.useDecimals ? 2 : 0
  };
  const formattedAmount = new Intl.NumberFormat(
    behavior.useComma ? 'de-DE' : 'en-US',
    options
  ).format(absAmount);
  const postfixCurrencies = new Set(['kr', 'kr.', 'Fr', 'zł']);
  const base = postfixCurrencies.has(behavior.symbol)
    ? `${formattedAmount} ${behavior.symbol}`
    : `${behavior.symbol}${formattedAmount}`;
  return isNegative ? `-${base}` : base;
}

export function getUserTimeZone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function formatMonth(date) {
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    timeZone: getUserTimeZone()
  });
}

export function getISODateWithLocalTime(dateInput) {
  const [year, month, day] = dateInput.split('-').map(Number);
  const now = new Date();
  const hours = now.getHours();
  const minutes = now.getMinutes();
  const seconds = now.getSeconds();
  const localDateTime = new Date(year, month - 1, day, hours, minutes, seconds);
  return localDateTime.toISOString();
}

export function formatDateFromUTC(utcDateString) {
  const date = new Date(utcDateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

export function getTodayBounds(date = new Date()) {
  const local = new Date(date);
  const startLocal = new Date(local.getFullYear(), local.getMonth(), local.getDate(), 0, 0, 0, 0);
  const endLocal = new Date(local.getFullYear(), local.getMonth(), local.getDate(), 23, 59, 59, 999);
  return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
}

export function getWeekBounds(date = new Date()) {
  const local = new Date(date);
  const day = local.getDay();
  const mondayIndex = (day + 6) % 7; // shift so Monday becomes start of week
  const startLocal = new Date(local.getFullYear(), local.getMonth(), local.getDate(), 0, 0, 0, 0);
  startLocal.setDate(startLocal.getDate() - mondayIndex);
  const endLocal = new Date(startLocal);
  endLocal.setDate(startLocal.getDate() + 6);
  endLocal.setHours(23, 59, 59, 999);
  return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
}

export function getMonthBounds(date, startDate = 1, endOfMonth = false) {
  const localDate = new Date(date);
  if (Number.isNaN(localDate.getTime())) {
    return { start: new Date(NaN), end: new Date(NaN) };
  }

  if (endOfMonth) {
    const startLocal = new Date(localDate.getFullYear(), localDate.getMonth(), 0);
    startLocal.setHours(0, 0, 0, 0);

    const endLocal = new Date(localDate.getFullYear(), localDate.getMonth() + 1, 0);
    endLocal.setDate(endLocal.getDate() - 1);
    endLocal.setHours(23, 59, 59, 999);

    return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
  }

  if (startDate === 1) {
    const startLocal = new Date(localDate.getFullYear(), localDate.getMonth(), 1);
    const endLocal = new Date(localDate.getFullYear(), localDate.getMonth() + 1, 0, 23, 59, 59, 999);
    return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
  }

  let thisMonthStartDate = Math.min(startDate, new Date(localDate.getFullYear(), localDate.getMonth() + 1, 0).getDate());
  const prevMonth = localDate.getMonth() === 0 ? 11 : localDate.getMonth() - 1;
  const prevYear = localDate.getMonth() === 0 ? localDate.getFullYear() - 1 : localDate.getFullYear();
  const daysInPrevMonth = new Date(prevYear, prevMonth + 1, 0).getDate();
  const prevMonthStartDate = Math.min(startDate, daysInPrevMonth);

  if (localDate.getDate() < thisMonthStartDate) {
    const startLocal = new Date(prevYear, prevMonth, prevMonthStartDate);
    const endLocal = new Date(localDate.getFullYear(), localDate.getMonth(), thisMonthStartDate - 1, 23, 59, 59, 999);
    return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
  }

  const nextMonth = localDate.getMonth() === 11 ? 0 : localDate.getMonth() + 1;
  const nextYear = localDate.getMonth() === 11 ? localDate.getFullYear() + 1 : localDate.getFullYear();
  const daysInNextMonth = new Date(nextYear, nextMonth + 1, 0).getDate();
  const nextMonthStartDate = Math.min(startDate, daysInNextMonth);
  const startLocal = new Date(localDate.getFullYear(), localDate.getMonth(), thisMonthStartDate);
  const endLocal = new Date(nextYear, nextMonth, nextMonthStartDate - 1, 23, 59, 59, 999);
  return { start: new Date(startLocal.toISOString()), end: new Date(endLocal.toISOString()) };
}

export function getCycleAnchor(date, startDate = 1, endOfMonth = false) {
  const localDate = new Date(date);
  if (Number.isNaN(localDate.getTime())) {
    return new Date(NaN);
  }

  if (endOfMonth) {
    const aligned = new Date(localDate.getFullYear(), localDate.getMonth(), 1);
    aligned.setHours(0, 0, 0, 0);
    return aligned;
  }

  const aligned = new Date(localDate);
  aligned.setHours(0, 0, 0, 0);

  if (startDate <= 1) {
    aligned.setDate(1);
    return aligned;
  }

  const daysInMonth = new Date(aligned.getFullYear(), aligned.getMonth() + 1, 0).getDate();
  const effectiveStart = Math.min(startDate, daysInMonth);

  if (aligned.getDate() >= effectiveStart) {
    aligned.setMonth(aligned.getMonth() + 1);
  }

  aligned.setDate(1);
  return aligned;
}

export function getMonthExpenses(expenses, currentDate, startDate, endOfMonth = false) {
  const { start, end } = getMonthBounds(currentDate, startDate, endOfMonth);
  return expenses
    .filter((exp) => {
      const expDate = new Date(exp.date);
      return expDate >= start && expDate <= end;
    })
    .sort((a, b) => new Date(b.date) - new Date(a.date));
}

export function filterExpensesByRange(expenses, range, referenceDate, startDate, endOfMonth = false) {
  let bounds;
  if (range === 'today') {
    bounds = getTodayBounds(referenceDate);
  } else if (range === 'week') {
    bounds = getWeekBounds(referenceDate);
  } else {
    bounds = getMonthBounds(referenceDate, startDate, endOfMonth);
  }
  const { start, end } = bounds;
  return expenses
    .filter((exp) => {
      const expDate = new Date(exp.date);
      return expDate >= start && expDate <= end;
    })
    .sort((a, b) => new Date(b.date) - new Date(a.date));
}

export function formatWeekRange(date = new Date()) {
  const { start, end } = getWeekBounds(date);
  const sameYear = start.getFullYear() === end.getFullYear();
  const baseOptions = { month: 'short', day: 'numeric' };
  const startLabel = start.toLocaleDateString('en-US', sameYear ? baseOptions : { ...baseOptions, year: 'numeric' });
  const endLabel = end.toLocaleDateString('en-US', { ...baseOptions, year: 'numeric' });
  return `${startLabel} - ${endLabel}`;
}

export function formatDayLabel(date = new Date()) {
  const today = new Date();
  const sameDay =
    today.getFullYear() === date.getFullYear() &&
    today.getMonth() === date.getMonth() &&
    today.getDate() === date.getDate();
  if (sameDay) {
    return 'Today';
  }
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

export function escapeHTML(str) {
  if (typeof str !== 'string') return str;
  return str.replace(/[&<>'"]/g, (tag) =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
  );
}
