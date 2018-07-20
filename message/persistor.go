package message

// Encodable is an object that can encode itself to a byte stream.
type Encodable interface {
	Encode(io.Writer) error
}

// Decodable is an object that can decode itself from a byte stream.
type Decodable interface {
	Decode(io.Reader) error
}

// FilePersistor allows one to save and load data from a file.
type FilePersistor struct {
	Path string
}

func (fp FilePersistor) Save(object Encodable) error {
	f, err := os.Create(fp.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return object.Encode(f)
}

func (fp FilePersistor) Load(object Decodable) error {
	f, err := os.Open(fp.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return object.Decode(f)
}

func MemPersistor struct {
	buf bytes.Buffer
}

func (mp MemPersistor) Save(object Encodable) error {
	return object.Encode(mp.buf)
}

func (mp MemPersistor) Load(object Decodable) error {
	return object.Decode(mp.buf)
}
