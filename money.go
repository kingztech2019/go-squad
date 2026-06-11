package squad

// NGN converts a naira amount to kobo (the lowest denomination used by the Squad API).
//
//	squad.NGN(5000)  // → 500000 kobo (₦5,000)
//	squad.NGN(1)     // → 100 kobo (₦1)
func NGN(naira float64) int64 {
	return int64(naira * 100)
}

// USD converts a dollar amount to cents (the lowest denomination for USD transactions).
//
//	squad.USD(50)   // → 5000 cents ($50.00)
//	squad.USD(0.50) // → 50 cents ($0.50)
func USD(dollars float64) int64 {
	return int64(dollars * 100)
}

// FromKobo converts a kobo amount back to naira for display purposes.
//
//	squad.FromKobo(500000) // → 5000.00 (₦5,000)
func FromKobo(kobo int64) float64 {
	return float64(kobo) / 100
}

// FromCents converts a cents amount back to dollars for display purposes.
//
//	squad.FromCents(5000) // → 50.00 ($50.00)
func FromCents(cents int64) float64 {
	return float64(cents) / 100
}
