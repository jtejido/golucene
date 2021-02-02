package standard

import (
	. "github.com/balzaczyy/golucene/analysis/util"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"io"
)

// standard/StandardTokenizerImpl.java

/* initial size of the lookahead buffer */
const ZZ_BUFFERSIZE = 255

/* lexical states */
const YYINITIAL = 0

/*
ZZ_LEXSTATE[l] is the state in the DFA for the lexical state l
ZZ_LEXSTATE[l+1] is the state in the DFA for the lexical state l at the beginning of a line
l is of the form l = 2*k, k a non negative integer
*/
var ZZ_LEXSTATE = [2]int{0, 0}

/* Translates characters to character classes */
var ZZ_CMAP_PACKED = []int{
	0042, 000, 001, 0015, 004, 000, 001, 0014, 004, 000, 001, 007, 001, 000, 001, 0010, 001, 000, 0012, 004,
	001, 006, 001, 007, 005, 000, 0032, 001, 004, 000, 001, 0011, 001, 000, 0032, 001, 0057, 000, 001, 001,
	002, 000, 001, 003, 007, 000, 001, 001, 001, 000, 001, 006, 002, 000, 001, 001, 005, 000, 0027, 001,
	001, 000, 0037, 001, 001, 000, int('\u01ca'), 001, 004, 000, 0014, 001, 005, 000, 001, 006, 0010, 000, 005, 001,
	007, 000, 001, 001, 001, 000, 001, 001, 0021, 000, 00160, 003, 005, 001, 001, 000, 002, 001, 002, 000,
	004, 001, 001, 007, 007, 000, 001, 001, 001, 006, 003, 001, 001, 000, 001, 001, 001, 000, 0024, 001,
	001, 000, 00123, 001, 001, 000, 00213, 001, 001, 000, 007, 003, 00236, 001, 0011, 000, 0046, 001, 002, 000,
	001, 001, 007, 000, 0047, 001, 001, 000, 001, 007, 007, 000, 0055, 003, 001, 000, 001, 003, 001, 000,
	002, 003, 001, 000, 002, 003, 001, 000, 001, 003, 0010, 000, 0033, 0016, 005, 000, 003, 0016, 001, 001,
	001, 006, 0013, 000, 005, 003, 007, 000, 002, 007, 002, 000, 0013, 003, 001, 000, 001, 003, 003, 000,
	0053, 001, 0025, 003, 0012, 004, 001, 000, 001, 004, 001, 007, 001, 000, 002, 001, 001, 003, 00143, 001,
	001, 000, 001, 001, 0010, 003, 001, 000, 006, 003, 002, 001, 002, 003, 001, 000, 004, 003, 002, 001,
	0012, 004, 003, 001, 002, 000, 001, 001, 0017, 000, 001, 003, 001, 001, 001, 003, 0036, 001, 0033, 003,
	002, 000, 00131, 001, 0013, 003, 001, 001, 0016, 000, 0012, 004, 0041, 001, 0011, 003, 002, 001, 002, 000,
	001, 007, 001, 000, 001, 001, 005, 000, 0026, 001, 004, 003, 001, 001, 0011, 003, 001, 001, 003, 003,
	001, 001, 005, 003, 0022, 000, 0031, 001, 003, 003, 00104, 000, 001, 001, 001, 000, 0013, 001, 0067, 000,
	0033, 003, 001, 000, 004, 003, 0066, 001, 003, 003, 001, 001, 0022, 003, 001, 001, 007, 003, 0012, 001,
	002, 003, 002, 000, 0012, 004, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 003, 003, 001, 000,
	0010, 001, 002, 000, 002, 001, 002, 000, 0026, 001, 001, 000, 007, 001, 001, 000, 001, 001, 003, 000,
	004, 001, 002, 000, 001, 003, 001, 001, 007, 003, 002, 000, 002, 003, 002, 000, 003, 003, 001, 001,
	0010, 000, 001, 003, 004, 000, 002, 001, 001, 000, 003, 001, 002, 003, 002, 000, 0012, 004, 002, 001,
	0017, 000, 003, 003, 001, 000, 006, 001, 004, 000, 002, 001, 002, 000, 0026, 001, 001, 000, 007, 001,
	001, 000, 002, 001, 001, 000, 002, 001, 001, 000, 002, 001, 002, 000, 001, 003, 001, 000, 005, 003,
	004, 000, 002, 003, 002, 000, 003, 003, 003, 000, 001, 003, 007, 000, 004, 001, 001, 000, 001, 001,
	007, 000, 0012, 004, 002, 003, 003, 001, 001, 003, 0013, 000, 003, 003, 001, 000, 0011, 001, 001, 000,
	003, 001, 001, 000, 0026, 001, 001, 000, 007, 001, 001, 000, 002, 001, 001, 000, 005, 001, 002, 000,
	001, 003, 001, 001, 0010, 003, 001, 000, 003, 003, 001, 000, 003, 003, 002, 000, 001, 001, 0017, 000,
	002, 001, 002, 003, 002, 000, 0012, 004, 0021, 000, 003, 003, 001, 000, 0010, 001, 002, 000, 002, 001,
	002, 000, 0026, 001, 001, 000, 007, 001, 001, 000, 002, 001, 001, 000, 005, 001, 002, 000, 001, 003,
	001, 001, 007, 003, 002, 000, 002, 003, 002, 000, 003, 003, 0010, 000, 002, 003, 004, 000, 002, 001,
	001, 000, 003, 001, 002, 003, 002, 000, 0012, 004, 001, 000, 001, 001, 0020, 000, 001, 003, 001, 001,
	001, 000, 006, 001, 003, 000, 003, 001, 001, 000, 004, 001, 003, 000, 002, 001, 001, 000, 001, 001,
	001, 000, 002, 001, 003, 000, 002, 001, 003, 000, 003, 001, 003, 000, 0014, 001, 004, 000, 005, 003,
	003, 000, 003, 003, 001, 000, 004, 003, 002, 000, 001, 001, 006, 000, 001, 003, 0016, 000, 0012, 004,
	0021, 000, 003, 003, 001, 000, 0010, 001, 001, 000, 003, 001, 001, 000, 0027, 001, 001, 000, 0012, 001,
	001, 000, 005, 001, 003, 000, 001, 001, 007, 003, 001, 000, 003, 003, 001, 000, 004, 003, 007, 000,
	002, 003, 001, 000, 002, 001, 006, 000, 002, 001, 002, 003, 002, 000, 0012, 004, 0022, 000, 002, 003,
	001, 000, 0010, 001, 001, 000, 003, 001, 001, 000, 0027, 001, 001, 000, 0012, 001, 001, 000, 005, 001,
	002, 000, 001, 003, 001, 001, 007, 003, 001, 000, 003, 003, 001, 000, 004, 003, 007, 000, 002, 003,
	007, 000, 001, 001, 001, 000, 002, 001, 002, 003, 002, 000, 0012, 004, 001, 000, 002, 001, 0017, 000,
	002, 003, 001, 000, 0010, 001, 001, 000, 003, 001, 001, 000, 0051, 001, 002, 000, 001, 001, 007, 003,
	001, 000, 003, 003, 001, 000, 004, 003, 001, 001, 0010, 000, 001, 003, 0010, 000, 002, 001, 002, 003,
	002, 000, 0012, 004, 0012, 000, 006, 001, 002, 000, 002, 003, 001, 000, 0022, 001, 003, 000, 0030, 001,
	001, 000, 0011, 001, 001, 000, 001, 001, 002, 000, 007, 001, 003, 000, 001, 003, 004, 000, 006, 003,
	001, 000, 001, 003, 001, 000, 0010, 003, 0022, 000, 002, 003, 0015, 000, 0060, 0020, 001, 0021, 002, 0020,
	007, 0021, 005, 000, 007, 0020, 0010, 0021, 001, 000, 0012, 004, 0047, 000, 002, 0020, 001, 000, 001, 0020,
	002, 000, 002, 0020, 001, 000, 001, 0020, 002, 000, 001, 0020, 006, 000, 004, 0020, 001, 000, 007, 0020,
	001, 000, 003, 0020, 001, 000, 001, 0020, 001, 000, 001, 0020, 002, 000, 002, 0020, 001, 000, 004, 0020,
	001, 0021, 002, 0020, 006, 0021, 001, 000, 002, 0021, 001, 0020, 002, 000, 005, 0020, 001, 000, 001, 0020,
	001, 000, 006, 0021, 002, 000, 0012, 004, 002, 000, 004, 0020, 0040, 000, 001, 001, 0027, 000, 002, 003,
	006, 000, 0012, 004, 0013, 000, 001, 003, 001, 000, 001, 003, 001, 000, 001, 003, 004, 000, 002, 003,
	0010, 001, 001, 000, 0044, 001, 004, 000, 0024, 003, 001, 000, 002, 003, 005, 001, 0013, 003, 001, 000,
	0044, 003, 0011, 000, 001, 003, 0071, 000, 0053, 0020, 0024, 0021, 001, 0020, 0012, 004, 006, 000, 006, 0020,
	004, 0021, 004, 0020, 003, 0021, 001, 0020, 003, 0021, 002, 0020, 007, 0021, 003, 0020, 004, 0021, 0015, 0020,
	0014, 0021, 001, 0020, 001, 0021, 0012, 004, 004, 0021, 002, 0020, 0046, 001, 001, 000, 001, 001, 005, 000,
	001, 001, 002, 000, 0053, 001, 001, 000, 004, 001, int('\u0100'), 002, 00111, 001, 001, 000, 004, 001, 002, 000,
	007, 001, 001, 000, 001, 001, 001, 000, 004, 001, 002, 000, 0051, 001, 001, 000, 004, 001, 002, 000,
	0041, 001, 001, 000, 004, 001, 002, 000, 007, 001, 001, 000, 001, 001, 001, 000, 004, 001, 002, 000,
	0017, 001, 001, 000, 0071, 001, 001, 000, 004, 001, 002, 000, 00103, 001, 002, 000, 003, 003, 0040, 000,
	0020, 001, 0020, 000, 00125, 001, 0014, 000, int('\u026c'), 001, 002, 000, 0021, 001, 001, 000, 0032, 001, 005, 000,
	00113, 001, 003, 000, 003, 001, 0017, 000, 0015, 001, 001, 000, 004, 001, 003, 003, 0013, 000, 0022, 001,
	003, 003, 0013, 000, 0022, 001, 002, 003, 0014, 000, 0015, 001, 001, 000, 003, 001, 001, 000, 002, 003,
	0014, 000, 0064, 0020, 0040, 0021, 003, 000, 001, 0020, 004, 000, 001, 0020, 001, 0021, 002, 000, 0012, 004,
	0041, 000, 004, 003, 001, 000, 0012, 004, 006, 000, 00130, 001, 0010, 000, 0051, 001, 001, 003, 001, 001,
	005, 000, 00106, 001, 0012, 000, 0035, 001, 003, 000, 0014, 003, 004, 000, 0014, 003, 0012, 000, 0012, 004,
	0036, 0020, 002, 000, 005, 0020, 0013, 000, 0054, 0020, 004, 000, 0021, 0021, 007, 0020, 002, 0021, 006, 000,
	0012, 004, 001, 0020, 003, 000, 002, 0020, 0040, 000, 0027, 001, 005, 003, 004, 000, 0065, 0020, 0012, 0021,
	001, 000, 0035, 0021, 002, 000, 001, 003, 0012, 004, 006, 000, 0012, 004, 006, 000, 0016, 0020, 00122, 000,
	005, 003, 0057, 001, 0021, 003, 007, 001, 004, 000, 0012, 004, 0021, 000, 0011, 003, 0014, 000, 003, 003,
	0036, 001, 0015, 003, 002, 001, 0012, 004, 0054, 001, 0016, 003, 0014, 000, 0044, 001, 0024, 003, 0010, 000,
	0012, 004, 003, 000, 003, 001, 0012, 004, 0044, 001, 00122, 000, 003, 003, 001, 000, 0025, 003, 004, 001,
	001, 003, 004, 001, 003, 003, 002, 001, 0011, 000, 00300, 001, 0047, 003, 0025, 000, 004, 003, int('\u0116'), 001,
	002, 000, 006, 001, 002, 000, 0046, 001, 002, 000, 006, 001, 002, 000, 0010, 001, 001, 000, 001, 001,
	001, 000, 001, 001, 001, 000, 001, 001, 001, 000, 0037, 001, 002, 000, 0065, 001, 001, 000, 007, 001,
	001, 000, 001, 001, 003, 000, 003, 001, 001, 000, 007, 001, 003, 000, 004, 001, 002, 000, 006, 001,
	004, 000, 0015, 001, 005, 000, 003, 001, 001, 000, 007, 001, 0017, 000, 004, 003, 0010, 000, 002, 0010,
	0012, 000, 001, 0010, 002, 000, 001, 006, 002, 000, 005, 003, 0020, 000, 002, 0011, 003, 000, 001, 007,
	0017, 000, 001, 0011, 0013, 000, 005, 003, 001, 000, 0012, 003, 001, 000, 001, 001, 0015, 000, 001, 001,
	0020, 000, 0015, 001, 0063, 000, 0041, 003, 0021, 000, 001, 001, 004, 000, 001, 001, 002, 000, 0012, 001,
	001, 000, 001, 001, 003, 000, 005, 001, 006, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001,
	001, 000, 004, 001, 001, 000, 0013, 001, 002, 000, 004, 001, 005, 000, 005, 001, 004, 000, 001, 001,
	0021, 000, 0051, 001, int('\u032d'), 000, 0064, 001, int('\u0716'), 000, 0057, 001, 001, 000, 0057, 001, 001, 000, 00205, 001,
	006, 000, 004, 001, 003, 003, 002, 001, 0014, 000, 0046, 001, 001, 000, 001, 001, 005, 000, 001, 001,
	002, 000, 0070, 001, 007, 000, 001, 001, 0017, 000, 001, 003, 0027, 001, 0011, 000, 007, 001, 001, 000,
	007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000,
	007, 001, 001, 000, 007, 001, 001, 000, 0040, 003, 0057, 000, 001, 001, 00120, 000, 0032, 0012, 001, 000,
	00131, 0012, 0014, 000, 00326, 0012, 0057, 000, 001, 001, 001, 000, 001, 0012, 0031, 000, 0011, 0012, 006, 003,
	001, 000, 005, 005, 002, 000, 003, 0012, 001, 001, 001, 001, 004, 000, 00126, 0013, 002, 000, 002, 003,
	002, 005, 003, 0013, 00133, 005, 001, 000, 004, 005, 005, 000, 0051, 001, 003, 000, 00136, 002, 0021, 000,
	0033, 001, 0065, 000, 0020, 005, 00320, 000, 0057, 005, 001, 000, 00130, 005, 00250, 000, int('\u19b6'), 0012, 00112, 000,
	int('\u51cd'), 0012, 0063, 000, int('\u048d'), 001, 00103, 000, 0056, 001, 002, 000, int('\u010d'), 001, 003, 000, 0020, 001, 0012, 004,
	002, 001, 0024, 000, 0057, 001, 004, 003, 001, 000, 0012, 003, 001, 000, 0031, 001, 007, 000, 001, 003,
	00120, 001, 002, 003, 0045, 000, 0011, 001, 002, 000, 00147, 001, 002, 000, 004, 001, 001, 000, 004, 001,
	0014, 000, 0013, 001, 00115, 000, 0012, 001, 001, 003, 003, 001, 001, 003, 004, 001, 001, 003, 0027, 001,
	005, 003, 0030, 000, 0064, 001, 0014, 000, 002, 003, 0062, 001, 0021, 003, 0013, 000, 0012, 004, 006, 000,
	0022, 003, 006, 001, 003, 000, 001, 001, 004, 000, 0012, 004, 0034, 001, 0010, 003, 002, 000, 0027, 001,
	0015, 003, 0014, 000, 0035, 002, 003, 000, 004, 003, 0057, 001, 0016, 003, 0016, 000, 001, 001, 0012, 004,
	0046, 000, 0051, 001, 0016, 003, 0011, 000, 003, 001, 001, 003, 0010, 001, 002, 003, 002, 000, 0012, 004,
	006, 000, 0033, 0020, 001, 0021, 004, 000, 0060, 0020, 001, 0021, 001, 0020, 003, 0021, 002, 0020, 002, 0021,
	005, 0020, 002, 0021, 001, 0020, 001, 0021, 001, 0020, 0030, 000, 005, 0020, 0013, 001, 005, 003, 002, 000,
	003, 001, 002, 003, 0012, 000, 006, 001, 002, 000, 006, 001, 002, 000, 006, 001, 0011, 000, 007, 001,
	001, 000, 007, 001, 00221, 000, 0043, 001, 0010, 003, 001, 000, 002, 003, 002, 000, 0012, 004, 006, 000,
	int('\u2ba4'), 002, 0014, 000, 0027, 002, 004, 000, 0061, 002, int('\u2104'), 000, int('\u016e'), 0012, 002, 000, 00152, 0012, 0046, 000,
	007, 001, 0014, 000, 005, 001, 005, 000, 001, 0016, 001, 003, 0012, 0016, 001, 000, 0015, 0016, 001, 000,
	005, 0016, 001, 000, 001, 0016, 001, 000, 002, 0016, 001, 000, 002, 0016, 001, 000, 0012, 0016, 00142, 001,
	0041, 000, int('\u016b'), 001, 0022, 000, 00100, 001, 002, 000, 0066, 001, 0050, 000, 0014, 001, 004, 000, 0020, 003,
	001, 007, 002, 000, 001, 006, 001, 007, 0013, 000, 007, 003, 0014, 000, 002, 0011, 0030, 000, 003, 0011,
	001, 007, 001, 000, 001, 0010, 001, 000, 001, 007, 001, 006, 0032, 000, 005, 001, 001, 000, 00207, 001,
	002, 000, 001, 003, 007, 000, 001, 0010, 004, 000, 001, 007, 001, 000, 001, 0010, 001, 000, 0012, 004,
	001, 006, 001, 007, 005, 000, 0032, 001, 004, 000, 001, 0011, 001, 000, 0032, 001, 0013, 000, 0070, 005,
	002, 003, 0037, 002, 003, 000, 006, 002, 002, 000, 006, 002, 002, 000, 006, 002, 002, 000, 003, 002,
	0034, 000, 003, 003, 004, 000, 0014, 001, 001, 000, 0032, 001, 001, 000, 0023, 001, 001, 000, 002, 001,
	001, 000, 0017, 001, 002, 000, 0016, 001, 0042, 000, 00173, 001, 00105, 000, 0065, 001, 00210, 000, 001, 003,
	00202, 000, 0035, 001, 003, 000, 0061, 001, 0057, 000, 0037, 001, 0021, 000, 0033, 001, 0065, 000, 0036, 001,
	002, 000, 0044, 001, 004, 000, 0010, 001, 001, 000, 005, 001, 0052, 000, 00236, 001, 002, 000, 0012, 004,
	int('\u0356'), 000, 006, 001, 002, 000, 001, 001, 001, 000, 0054, 001, 001, 000, 002, 001, 003, 000, 001, 001,
	002, 000, 0027, 001, 00252, 000, 0026, 001, 0012, 000, 0032, 001, 00106, 000, 0070, 001, 006, 000, 002, 001,
	00100, 000, 001, 001, 003, 003, 001, 000, 002, 003, 005, 000, 004, 003, 004, 001, 001, 000, 003, 001,
	001, 000, 0033, 001, 004, 000, 003, 003, 004, 000, 001, 003, 0040, 000, 0035, 001, 00203, 000, 0066, 001,
	0012, 000, 0026, 001, 0012, 000, 0023, 001, 00215, 000, 00111, 001, int('\u03b7'), 000, 003, 003, 0065, 001, 0017, 003,
	0037, 000, 0012, 004, 0020, 000, 003, 003, 0055, 001, 0013, 003, 002, 000, 001, 003, 0022, 000, 0031, 001,
	007, 000, 0012, 004, 006, 000, 003, 003, 0044, 001, 0016, 003, 001, 000, 0012, 004, 00100, 000, 003, 003,
	0060, 001, 0016, 003, 004, 001, 0013, 000, 0012, 004, int('\u04a6'), 000, 0053, 001, 0015, 003, 0010, 000, 0012, 004,
	int('\u0936'), 000, int('\u036f'), 001, 00221, 000, 00143, 001, int('\u0b9d'), 000, int('\u042f'), 001, int('\u33d1'), 000, int('\u0239'), 001, int('\u04c7'), 000, 00105, 001,
	0013, 000, 001, 001, 0056, 003, 0020, 000, 004, 003, 0015, 001, int('\u4060'), 000, 001, 005, 001, 0013, int('\u2163'), 000,
	005, 003, 003, 000, 0026, 003, 002, 000, 007, 003, 0036, 000, 004, 003, 00224, 000, 003, 003, int('\u01bb'), 000,
	00125, 001, 001, 000, 00107, 001, 001, 000, 002, 001, 002, 000, 001, 001, 002, 000, 002, 001, 002, 000,
	004, 001, 001, 000, 0014, 001, 001, 000, 001, 001, 001, 000, 007, 001, 001, 000, 00101, 001, 001, 000,
	004, 001, 002, 000, 0010, 001, 001, 000, 007, 001, 001, 000, 0034, 001, 001, 000, 004, 001, 001, 000,
	005, 001, 001, 000, 001, 001, 003, 000, 007, 001, 001, 000, int('\u0154'), 001, 002, 000, 0031, 001, 001, 000,
	0031, 001, 001, 000, 0037, 001, 001, 000, 0031, 001, 001, 000, 0037, 001, 001, 000, 0031, 001, 001, 000,
	0037, 001, 001, 000, 0031, 001, 001, 000, 0037, 001, 001, 000, 0031, 001, 001, 000, 0010, 001, 002, 000,
	0062, 004, int('\u1600'), 000, 004, 001, 001, 000, 0033, 001, 001, 000, 002, 001, 001, 000, 001, 001, 002, 000,
	001, 001, 001, 000, 0012, 001, 001, 000, 004, 001, 001, 000, 001, 001, 001, 000, 001, 001, 006, 000,
	001, 001, 004, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000, 003, 001, 001, 000,
	002, 001, 001, 000, 001, 001, 002, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000,
	001, 001, 001, 000, 001, 001, 001, 000, 002, 001, 001, 000, 001, 001, 002, 000, 004, 001, 001, 000,
	007, 001, 001, 000, 004, 001, 001, 000, 004, 001, 001, 000, 001, 001, 001, 000, 0012, 001, 001, 000,
	0021, 001, 005, 000, 003, 001, 001, 000, 005, 001, 001, 000, 0021, 001, int('\u032a'), 000, 0032, 0017, 001, 0013,
	int('\u0dff'), 000, int('\ua6d7'), 0012, 0051, 000, int('\u1035'), 0012, 0013, 000, 00336, 0012, int('\u3fe2'), 000, int('\u021e'), 0012, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\u05ee'), 000,
	001, 003, 0036, 000, 00140, 003, 00200, 000, 00360, 003, int('\uffff'), 000, int('\uffff'), 000, int('\ufe12'), 000,
}

/* Translates characters to character classes */
var ZZ_CMAP = zzUnpackCMap(ZZ_CMAP_PACKED)

/* Unpacks the comressed character translation table */
func zzUnpackCMap(packed []int) []rune {
	m := make([]rune, 0x110000)
	j := 0 // index in unpacked array
	assert(len(packed) == 2836)

	for i := 0; i < 2836; i += 2 {
		count, value := packed[i], packed[i+1]
		m[j] = rune(value)
		j++
		count--
		for count > 0 {
			m[j] = rune(value)
			j++
			count--
		}
	}
	return m
}

/* Translates DFA states to action switch labels. */
var ZZ_ACTION = zzUnpackAction([]int{
	001, 000, 001, 001, 001, 002, 001, 003, 001, 004, 001, 005, 001, 001, 001, 006,
	001, 007, 001, 002, 001, 001, 001, 0010, 001, 002, 001, 000, 001, 002, 001, 000,
	001, 004, 001, 000, 002, 002, 002, 000, 001, 001, 001, 000,
})

func zzUnpackAction(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
	}
	return m
}

/* Translates a state to a row index in the transition table */
var ZZ_ROWMAP = zzUnpackRowMap([]int{
	000, 000, 000, 0022, 000, 0044, 000, 0066, 000, 00110, 000, 00132, 000, 00154, 000, 00176,
	000, 00220, 000, 00242, 000, 00264, 000, 00306, 000, 00330, 000, 00352, 000, 00374, 000, int('\u010e'),
	000, int('\u0120'), 000, 00154, 000, int('\u0132'), 000, int('\u0144'), 000, int('\u0156'), 000, 00264, 000, int('\u0168'), 000, int('\u017a'),
})

func zzUnpackRowMap(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		high, low := packed[i]<<16, packed[i+1]
		m[j] = high | low
		j++
	}
	return m
}

/* The transition table of the DFA */
var ZZ_TRANS = zzUnpackTrans([]int{
	001, 002, 001, 003, 001, 004, 001, 002, 001, 005, 001, 006, 003, 002, 001, 007,
	001, 0010, 001, 0011, 002, 002, 001, 0012, 001, 0013, 002, 0014, 0023, 000, 003, 003,
	001, 0015, 001, 000, 001, 0016, 001, 000, 001, 0016, 001, 0017, 002, 000, 001, 0016,
	001, 000, 001, 0012, 002, 000, 001, 003, 001, 000, 001, 003, 002, 004, 001, 0015,
	001, 000, 001, 0016, 001, 000, 001, 0016, 001, 0017, 002, 000, 001, 0016, 001, 000,
	001, 0012, 002, 000, 001, 004, 001, 000, 002, 003, 002, 005, 002, 000, 002, 0020,
	001, 0021, 002, 000, 001, 0020, 001, 000, 001, 0012, 002, 000, 001, 005, 003, 000,
	001, 006, 001, 000, 001, 006, 003, 000, 001, 0017, 007, 000, 001, 006, 001, 000,
	002, 003, 001, 0022, 001, 005, 001, 0023, 003, 000, 001, 0022, 004, 000, 001, 0012,
	002, 000, 001, 0022, 003, 000, 001, 0010, 0015, 000, 001, 0010, 003, 000, 001, 0011,
	0015, 000, 001, 0011, 001, 000, 002, 003, 001, 0012, 001, 0015, 001, 000, 001, 0016,
	001, 000, 001, 0016, 001, 0017, 002, 000, 001, 0024, 001, 0025, 001, 0012, 002, 000,
	001, 0012, 003, 000, 001, 0026, 0013, 000, 001, 0027, 001, 000, 001, 0026, 003, 000,
	001, 0014, 0014, 000, 002, 0014, 001, 000, 002, 003, 002, 0015, 002, 000, 002, 0030,
	001, 0017, 002, 000, 001, 0030, 001, 000, 001, 0012, 002, 000, 001, 0015, 001, 000,
	002, 003, 001, 0016, 0012, 000, 001, 003, 002, 000, 001, 0016, 001, 000, 002, 003,
	001, 0017, 001, 0015, 001, 0023, 003, 000, 001, 0017, 004, 000, 001, 0012, 002, 000,
	001, 0017, 003, 000, 001, 0020, 001, 005, 0014, 000, 001, 0020, 001, 000, 002, 003,
	001, 0021, 001, 005, 001, 0023, 003, 000, 001, 0021, 004, 000, 001, 0012, 002, 000,
	001, 0021, 003, 000, 001, 0023, 001, 000, 001, 0023, 003, 000, 001, 0017, 007, 000,
	001, 0023, 001, 000, 002, 003, 001, 0024, 001, 0015, 004, 000, 001, 0017, 004, 000,
	001, 0012, 002, 000, 001, 0024, 003, 000, 001, 0025, 0012, 000, 001, 0024, 002, 000,
	001, 0025, 003, 000, 001, 0027, 0013, 000, 001, 0027, 001, 000, 001, 0027, 003, 000,
	001, 0030, 001, 0015, 0014, 000, 001, 0030,
})

func zzUnpackTrans(packed []int) []int {
	m := make([]int, 396)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]-1
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
	}
	return m
}

/* error codes */
const (
	ZZ_UNKNOWN_ERROR = 0
	ZZ_NO_MATCH      = 1
	ZZ_PUSHBACK_2BIG = 2
)

/* error messages for the codes above */
var ZZ_ERROR_MSG = [3]string{
	"Unkown internal scanner error",
	"Error: could not match input",
	"Error: pushback value was too large",
}

/* ZZ_ATTRIBUTE[aState] contains the attributes of state aState */
var ZZ_ATTRIBUTE = zzUnpackAttribute([]int{
	001, 000, 001, 011, 013, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000,
	002, 001, 002, 000, 001, 001, 001, 000,
})

func zzUnpackAttribute(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
		i += 2
	}
	return m
}

const (
	WORD_TYPE             = ALPHANUM
	NUMERIC_TYPE          = NUM
	SOUTH_EAST_ASIAN_TYPE = SOUTHEAST_ASIAN
	IDEOGRAPHIC_TYPE      = IDEOGRAPHIC
	HIRAGANA_TYPE         = HIRAGANA
	KATAKANA_TYPE         = KATAKANA
	HANGUL_TYPE           = HANGUL
)

/*
This class implements Word Break rules from the Unicode Text
Segmentation algorithm, as specified in Unicode Standard Annex #29.

Tokens produced are of the following types:

	- <ALPHANUM>: A sequence of alphabetic and numeric characters
	- <NUM>: A number
	- <SOUTHEAST_ASIAN>: A sequence of characters from South and Southeast Asian languages, including Thai, Lao, Myanmar, and Khmer
	- IDEOGRAPHIC>: A single CJKV ideographic character
	- <HIRAGANA>: A single hiragana character

Technically it should auto generated by JFlex but there is no GoFlex
yet. So it's a line-by-line port.
*/
type StandardTokenizerImpl struct {
	// the input device
	zzReader io.RuneReader

	// the current state of the DFA
	zzState int

	// the current lexical state
	zzLexicalState int

	// this buffer contains the current text to be matched and is the
	// source of yytext() string
	zzBuffer []rune

	// the text position at the last accepting state
	zzMarkedPos int

	// the current text position in the buffer
	zzCurrentPos int

	// startRead marks the beginning of the yytext() string in the buffer
	zzStartRead int

	// endRead marks the last character in the buffer, that has been read from input
	zzEndRead int

	// number of newlines encountered up to the start of the matched text
	yyline int

	// the number of characters up to the start of the matched text
	_yychar int

	// the number of characters from the last newline up to the start of the matched text
	yycolumn int

	// zzAtBOL == true <=> the scanner is currently at the beginning of a line
	zzAtBOL bool

	// zzAtEOF == true <=> the scanner is at the EOF
	zzAtEOF bool

	// denotes if the user-EOF-code has already been executed
	zzEOFDone bool

	// The number of occupied positions in zzBuffer beyond zzEndRead.
	// When a lead/high surrogate has been read from the input stream
	// into the final zzBuffer position, this will have a value of 1;
	// otherwise, it will have a value of 0.
	zzFinalHighSurrogate int
}

func newStandardTokenizerImpl(in io.RuneReader) *StandardTokenizerImpl {
	return &StandardTokenizerImpl{
		zzReader:       in,
		zzLexicalState: YYINITIAL,
		zzBuffer:       make([]rune, ZZ_BUFFERSIZE),
		zzAtBOL:        true,
	}
}

func (t *StandardTokenizerImpl) yychar() int {
	return t._yychar
}

/* Fills CharTermAttribute with the current token text. */
func (t *StandardTokenizerImpl) text(tt CharTermAttribute) {
	tt.CopyBuffer(t.zzBuffer[t.zzStartRead:t.zzMarkedPos])
}

/* Refills the input buffer. */
func (t *StandardTokenizerImpl) zzRefill() (bool, error) {
	// first: make room (if you can)
	if t.zzStartRead > 0 {
		t.zzEndRead += t.zzFinalHighSurrogate
		t.zzFinalHighSurrogate = 0
		//copy(t.zzBuffer[:t.zzEndRead-t.zzStartRead], t.zzBuffer[t.zzStartRead:t.zzEndRead-t.zzStartRead])
		copy(t.zzBuffer, t.zzBuffer[t.zzStartRead:t.zzEndRead])
		// translate stored positions
		t.zzEndRead -= t.zzStartRead
		t.zzCurrentPos -= t.zzStartRead
		t.zzMarkedPos -= t.zzStartRead
		t.zzStartRead = 0
	}

	// fill the buffer with new input
	var requested = len(t.zzBuffer) - t.zzEndRead - t.zzFinalHighSurrogate
	var totalRead = 0
	var err error
	for totalRead < requested && err != io.EOF {
		var numRead int
		if numRead, err = readRunes(t.zzReader.(io.RuneReader), t.zzBuffer[t.zzEndRead+totalRead:]); err != nil && err != io.EOF {
			return false, err
		}
		if numRead == -1 {
			break
		}
		totalRead += numRead
	}

	if totalRead > 0 {
		t.zzEndRead += totalRead
		if totalRead == requested { // possibly more input available
			if IsHighSurrogate(t.zzBuffer[t.zzEndRead-1]) {
				t.zzEndRead--
				t.zzFinalHighSurrogate = 1
				if totalRead == 1 {
					return true, nil
				}
			}
		}
		return false, nil
	}

	assert(totalRead == 0 && err == io.EOF)
	return true, nil
}

func readRunes(r io.RuneReader, buffer []rune) (int, error) {
	for i, _ := range buffer {
		ch, _, err := r.ReadRune()
		if err != nil {
			return i, err
		}
		buffer[i] = ch
	}
	return len(buffer), nil
}

/*
Resets the scanner to read from a new input stream.
Does not close the old reader.

All internal variables are reset, the old input stream
cannot be reused (internal buffer is discarded and lost).
Lexical state is set to ZZ_INITIAL.

Internal scan buffer is resized down to its initial length, if it has grown.
*/
func (t *StandardTokenizerImpl) yyreset(reader io.RuneReader) {
	t.zzReader = reader
	t.zzAtBOL = true
	t.zzAtEOF = false
	t.zzEOFDone = false
	t.zzEndRead, t.zzStartRead = 0, 0
	t.zzCurrentPos, t.zzMarkedPos = 0, 0
	t.zzFinalHighSurrogate = 0
	t.yyline, t._yychar, t.yycolumn = 0, 0, 0
	t.zzLexicalState = YYINITIAL
	if len(t.zzBuffer) > ZZ_BUFFERSIZE {
		t.zzBuffer = make([]rune, ZZ_BUFFERSIZE)
	}
}

/* Returns the length of the matched text region. */
func (t *StandardTokenizerImpl) yylength() int {
	return t.zzMarkedPos - t.zzStartRead
}

/*
Reports an error that occurred while scanning.

In a wellcormed scanner (no or only correct usage of yypushack(int)
and a match-all fallback rule) this method will only be called with
things that "can't possibly happen". If thismethod is called,
something is seriously wrong (e.g. a JFlex bug producing a faulty
scanner etc.).

Usual syntax/scanner level error handling should be done in error
fallback rules.
*/
func (t *StandardTokenizerImpl) zzScanError(errorCode int) {
	var msg string
	if errorCode >= 0 && errorCode < len(ZZ_ERROR_MSG) {
		msg = ZZ_ERROR_MSG[errorCode]
	} else {
		msg = ZZ_ERROR_MSG[ZZ_UNKNOWN_ERROR]
	}
	panic(msg)
}

/*
Resumes scanning until the next regular expression is matched, the
end of input is encountered or an I/O-Error occurs.
*/
func (t *StandardTokenizerImpl) nextToken() (int, error) {
	var zzInput, zzAction int

	// cached fields:
	var zzCurrentPosL, zzMarkedPosL int
	zzEndReadL := t.zzEndRead
	zzBufferL := t.zzBuffer
	zzCMapL := ZZ_CMAP

	zzTransL := ZZ_TRANS
	zzRowMapL := ZZ_ROWMAP
	zzAttrL := ZZ_ATTRIBUTE

	for {
		zzMarkedPosL = t.zzMarkedPos

		t._yychar += zzMarkedPosL - t.zzStartRead

		zzAction = -1

		zzCurrentPosL = zzMarkedPosL
		t.zzCurrentPos = zzMarkedPosL
		t.zzStartRead = zzMarkedPosL

		t.zzState = ZZ_LEXSTATE[t.zzLexicalState]

		// set up zzAction for empty match case:
		zzAttributes := zzAttrL[t.zzState]
		if (zzAttributes & 1) == 1 {
			zzAction = t.zzState
		}

	out:
		for {
			if zzCurrentPosL < zzEndReadL {
				// zzInput = int(zzBufferL[zzCurrentPosL])
				zzInput = CodePointAt(zzBufferL, zzCurrentPosL, zzEndReadL)

				zzCurrentPosL += CharCount(zzInput)
			} else if t.zzAtEOF {
				zzInput = YYEOF
				break out
			} else {
				// store back cached positions
				t.zzCurrentPos = zzCurrentPosL
				t.zzMarkedPos = zzMarkedPosL
				eof, err := t.zzRefill()
				if err != nil {
					return 0, err
				}
				// get translated positions and possibly new buffer
				zzCurrentPosL = t.zzCurrentPos
				zzMarkedPosL = t.zzMarkedPos
				zzBufferL = t.zzBuffer
				zzEndReadL = t.zzEndRead
				if eof {
					zzInput = YYEOF
					break out
				} else {
					zzInput = CodePointAt(zzBufferL, zzCurrentPosL, zzEndReadL)
					zzCurrentPosL += CharCount(zzInput)
				}
			}

			zzNext := zzTransL[zzRowMapL[t.zzState]+int(zzCMapL[zzInput])]
			if zzNext == -1 {
				break out
			}
			t.zzState = zzNext

			zzAttributes := zzAttrL[t.zzState]
			if (zzAttributes & 1) == 1 {
				zzAction = t.zzState
				zzMarkedPosL = zzCurrentPosL
				if (zzAttributes & 8) == 8 {
					break out
				}
			}
		}

		// store back cached position
		t.zzMarkedPos = zzMarkedPosL

		cond := zzAction
		if zzAction >= 0 {
			cond = ZZ_ACTION[zzAction]
		}

		switch cond {
		case 1, 9, 10, 11, 12, 13, 14, 15, 16:
			// break so we don't hit fall-through warning:
			// not numeric, word, ideographic, hiragana, or SE Asian -- ignore it.
			break
		case 2:
			return WORD_TYPE, nil
		case 3:
			return HANGUL_TYPE, nil
		case 4:
			return NUMERIC_TYPE, nil
		case 5:
			return KATAKANA_TYPE, nil
		case 6:
			return IDEOGRAPHIC_TYPE, nil
		case 7:
			return HIRAGANA_TYPE, nil
		case 8:
			return SOUTH_EAST_ASIAN_TYPE, nil
		default:
			if zzInput == YYEOF && t.zzStartRead == t.zzCurrentPos {
				t.zzAtEOF = true
				return YYEOF, nil
			} else {
				t.zzScanError(ZZ_NO_MATCH)
				panic("should not be here")
			}
		}
	}
}
