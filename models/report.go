package models

type AttestationReport struct {
	TestString string `json:"testString"`
}

//todo sinnvolle Struktur hinzufügen
/*
			Version
			Source
			Protocol
			Instance

				Attestation:
				Version
				Product
				Report
				Data

		Vorher oder nachher enthashen?

	Beide Reports enthalten unterschiedliche Attribute
*/
