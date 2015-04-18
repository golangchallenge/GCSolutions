package main

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"io/ioutil"
	"net"
	"testing"
)

func TestReadWriterPing(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, 1024)
	n, err := secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]

	// Make sure we have hello world back
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestSecureWriter(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, priv, pub)

	// Make sure we are secure
	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	if res := string(buf); res == "hello world\n" {
		t.Fatal("Unexpected result. The message is not encrypted.")
	}

	r, w = io.Pipe()
	secureW = NewSecureWriter(w, priv, pub)

	// Make sure we are unique
	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf2, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	if string(buf) == string(buf2) {
		t.Fatal("Unexpected result. The encrypted message is not unique.")
	}

}

func TestSecureEchoServer(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go Serve(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if got := string(buf[:n]); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
	}
}

func TestSecureServe(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	unexpected := "hello world\n"
	if _, err := fmt.Fprintf(conn, unexpected); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got == unexpected {
		t.Fatalf("Unexpected result:\nGot raw data instead of serialized key")
	}
}

func TestSecureDial(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				key := [32]byte{}
				c.Write(key[:])
				buf := make([]byte, 2048)
				n, err := c.Read(buf)
				if err != nil {
					t.Fatal(err)
				}
				if got := string(buf[:n]); got == "hello world\n" {
					t.Fatal("Unexpected result. Got raw data instead of encrypted")
				}
			}(conn)
		}
	}(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}
}

//
// Extra tests
//

func TestReadWriterMultiPing(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	// Encrypt hello world
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Fprintf(secureW, "hello world %d\n", i)
		}
		w.Close()
	}()

	buf, err := ioutil.ReadAll(secureR)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we have hello world back
	expected := "hello world 0\nhello world 1\nhello world 2\nhello world 3\nhello world 4\nhello world 5\nhello world 6\nhello world 7\nhello world 8\nhello world 9\n"
	if res := string(buf); res != expected {
		t.Fatalf("Unexpected result: %s != %s", res, expected)
	}
}

func TestAsymmetricalDecryptionWithBox(t *testing.T) {
	cpub, cpriv, _ := box.GenerateKey(rand.Reader)
	spub, spriv, _ := box.GenerateKey(rand.Reader)

	nonce := &[24]byte{'a'}
	message := []byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '\n'}

	encrypted := box.Seal([]byte{}, message, nonce, spub, cpriv)
	buf, _ := box.Open([]byte{}, encrypted, nonce, cpub, spriv)

	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestAsymmetricalDecryption(t *testing.T) {
	cpub, cpriv, _ := box.GenerateKey(rand.Reader)
	spub, spriv, _ := box.GenerateKey(rand.Reader)

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, cpriv, spub)
	secureR := NewSecureReader(r, spriv, cpub)

	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, 1024)
	n, err := secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]

	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestAsymmetricalDecryptionEcho(t *testing.T) {
	cpub, cpriv, _ := box.GenerateKey(rand.Reader)
	spub, spriv, _ := box.GenerateKey(rand.Reader)

	upR, upW := io.Pipe()
	downR, downW := io.Pipe()

	secureCW := NewSecureWriter(upW, cpriv, spub)
	secureCR := NewSecureReader(downR, cpriv, spub)

	secureSW := NewSecureWriter(downW, spriv, cpub)
	secureSR := NewSecureReader(upR, spriv, cpub)

	go func() {
		_, err := io.Copy(secureSW, secureSR)
		if err != nil {
			t.Fatal(err)
		}
		downW.Close()
	}()

	go func() {
		fmt.Fprintf(secureCW, "hello world\n")
		fmt.Fprintf(secureCW, "hello world2\n")
		upW.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf, err := ioutil.ReadAll(secureCR)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	expected := "hello world\nhello world2\n"
	if got := string(buf); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
	}
}

func TestLargeMessage(t *testing.T) {
	cpub, cpriv, _ := box.GenerateKey(rand.Reader)
	spub, spriv, _ := box.GenerateKey(rand.Reader)

	upR, upW := io.Pipe()
	downR, downW := io.Pipe()

	secureCW := NewSecureWriter(upW, cpriv, spub)
	secureCR := NewSecureReader(downR, cpriv, spub)

	secureSW := NewSecureWriter(downW, spriv, cpub)
	secureSR := NewSecureReader(upR, spriv, cpub)

	go func() {
		_, err := io.Copy(secureSW, secureSR)
		if err != nil {
			t.Fatal(err)
		}
		downW.Close()
	}()

	message := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aliquam fringilla mauris non lectus tempus congue. Quisque lobortis turpis eleifend ex mollis, at tempor eros lobortis. Ut id lectus pretium, consequat felis non, ornare velit. Fusce orci sapien, mattis sit amet euismod et, hendrerit at sem. In hac habitasse platea dictumst. Aliquam erat volutpat. Duis leo est, pharetra vitae commodo vitae, facilisis sed eros.

Mauris posuere libero a dui porttitor, vitae molestie felis ullamcorper. Nam et hendrerit ante. Vivamus magna ligula, consectetur eget magna vel, malesuada varius augue. In finibus sagittis mauris, et viverra tellus blandit eget. Nunc diam odio, blandit eu aliquet vitae, euismod et lectus. Pellentesque sodales pellentesque velit. Proin at augue dui. Integer blandit in massa in sagittis. Nullam lobortis posuere felis vel fermentum. Aliquam quam urna, tincidunt in molestie cursus, accumsan non enim. Nam dapibus euismod mauris, sit amet vehicula augue suscipit nec.

Proin sit amet eros nec nisi rutrum tristique. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Morbi eget hendrerit diam, at aliquet arcu. Fusce consectetur tincidunt hendrerit. Pellentesque elementum porta magna, vitae vulputate metus rhoncus non. Phasellus a lacinia orci, vitae consequat dui. Nulla eget erat in risus elementum malesuada non ultricies urna.

Mauris sed tellus mi. Integer eget risus id ante mollis viverra ultrices eget mi. Praesent eu efficitur lacus. Proin sem magna, tempor vel cursus in, pretium non arcu. Nunc vehicula convallis consectetur. Nullam in vulputate augue. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Pellentesque sed ullamcorper urna. Integer sed tortor ultricies, tempus mi a, rutrum ligula.

Nunc feugiat efficitur odio nec congue. Mauris vitae nunc porttitor, mattis nisl vel, tempor ligula. Donec cursus vehicula tincidunt. Maecenas finibus urna ut turpis aliquet auctor. Aliquam dapibus nec enim id pretium. Sed vel leo et velit molestie maximus. Nunc rhoncus, arcu non interdum pellentesque, nisl urna sodales elit, tristique tempus tellus nunc placerat nunc. Nullam iaculis porta tempor. Suspendisse velit nisl, vestibulum vel lobortis a, malesuada in urna. Proin id vestibulum ligula, sed luctus neque. Donec quis erat imperdiet, varius metus nec, tristique massa.

Maecenas tristique lectus vitae lacus lacinia, id rutrum dui mattis. Nullam imperdiet tellus nec malesuada maximus. Nulla non metus sed dolor hendrerit maximus. Vestibulum id malesuada felis. Duis ut velit sem. Praesent et volutpat turpis. Cras ac ex in nunc vulputate vehicula at vel sapien. Curabitur et vulputate felis. Cras ac bibendum libero. Aliquam eu mi non felis sodales congue id a quam. Nam auctor tempus felis sed fringilla. Sed vitae elementum ligula. Maecenas ornare augue at turpis tempus aliquam. Curabitur posuere ac nunc nec commodo. Lorem ipsum dolor sit amet, consectetur adipiscing elit.

Vivamus quis purus a metus pellentesque sagittis ac eu ipsum. Praesent id nisl scelerisque, pretium purus ac, consequat lorem. Maecenas ut felis dapibus, sagittis dolor vel, porta odio. Interdum et malesuada fames ac ante ipsum primis in faucibus. Proin sit amet erat et arcu aliquam efficitur. Proin arcu felis, dictum vel quam a, ultrices ultricies magna. Donec at tincidunt dui. Vivamus semper tellus sed tortor aliquet, id auctor nibh aliquet. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Cras vel tincidunt risus. Morbi mattis nunc leo, commodo iaculis purus cursus id.

Quisque tristique ex eget urna varius dapibus. Nullam pellentesque vitae odio sit amet commodo. Phasellus gravida egestas ligula. Nulla bibendum quis metus eu cursus. Praesent magna nunc, cursus in faucibus quis, volutpat efficitur nisl. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras egestas elit sit amet purus rutrum sagittis. Etiam gravida nulla urna, sed tincidunt urna pretium eget. Morbi sed sapien ut diam accumsan dictum. Cras mi risus, euismod ut elit molestie, aliquet malesuada arcu. Vestibulum rhoncus luctus tortor nec semper. Vivamus ante dui, mattis nec semper rhoncus, tincidunt quis dui. Vestibulum vitae aliquet libero. Donec faucibus rhoncus ex vitae lacinia. Aenean nunc purus, ultricies ac congue porta, sagittis a elit. In commodo est sem.

Donec feugiat consectetur sem eu lobortis. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Proin faucibus pharetra lorem iaculis dignissim. Mauris tempus tortor sit amet molestie rutrum. Vestibulum feugiat sollicitudin metus sed finibus. Curabitur dictum sollicitudin lorem, vitae faucibus nisi sagittis et. Vivamus at consectetur nunc. Aliquam sit amet massa ac ante efficitur pharetra. Nam tempor accumsan ligula nec mattis. Aenean risus ex, tincidunt quis fringilla vitae, faucibus sed mi. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. In hac habitasse platea dictumst. Curabitur nec dictum arcu, sed rutrum massa. Morbi urna mauris, finibus nec sapien quis, rhoncus tincidunt quam. Aenean metus enim, egestas varius facilisis ac, commodo ac libero.

Quisque nec aliquet dui, ac blandit turpis. Donec venenatis tellus non quam aliquam, mollis auctor justo dignissim. Duis dolor arcu, pellentesque eget nisi luctus, dapibus ullamcorper justo. Mauris molestie tellus at tortor consequat, in rhoncus velit ultricies. In non facilisis leo. Donec nec pharetra felis. Maecenas pharetra, tortor quis egestas fringilla, turpis mauris fringilla tellus, a placerat orci nunc eu sapien. Cras eget tortor dui.

Vivamus interdum imperdiet imperdiet. Ut libero dolor, fermentum sed nisl vitae, consectetur convallis leo. Maecenas finibus lacus et diam molestie vehicula. Morbi sed diam elit. Etiam urna lectus, laoreet a risus sit amet, aliquam mattis tellus. Ut eget pretium justo, vitae accumsan nunc. Suspendisse vel neque sit amet felis blandit egestas. Mauris quam tellus, rhoncus a est ac, auctor egestas odio. Ut vitae dignissim lacus. Vestibulum blandit dolor dolor, non consequat tortor cursus mattis. Ut libero elit, eleifend lobortis semper et, porttitor sed ligula. In sed nisi a urna efficitur consequat nec sed orci.

Quisque ut mauris quis lacus placerat cursus nec quis metus. Aenean lobortis justo congue nulla laoreet, ac scelerisque velit hendrerit. Quisque eros erat, suscipit sit amet erat ut, lacinia tempus metus. Nam venenatis elit sed mauris fermentum, eu auctor ante consectetur. Ut vehicula hendrerit est non varius. Vivamus rhoncus nisi ut purus porta scelerisque. Integer dictum tellus vel eros dictum, ut fermentum augue laoreet. Vivamus justo ex, pretium non vestibulum eu, tempus quis leo. Fusce dapibus condimentum faucibus.

Duis non euismod tellus. Sed interdum diam at tortor interdum pharetra. Nulla bibendum feugiat purus eu tempus. Sed eu tellus nec leo iaculis sagittis vitae et ipsum. Curabitur nec sollicitudin lorem. Sed ornare, eros sit amet semper sagittis, odio erat finibus ex, vitae placerat elit felis a arcu. In nisi turpis, viverra quis leo nec, ullamcorper egestas ex. Morbi et lorem sapien. Pellentesque luctus sem dapibus nisl placerat, sed rutrum nibh facilisis. Mauris lorem enim, egestas vitae rutrum non, aliquet in urna.

Morbi sit amet odio sed justo sagittis fermentum vitae a mi. Integer non est sit amet dui eleifend lobortis vel vel dui. Sed bibendum nulla ac est aliquam, nec egestas risus facilisis. Aenean ac ex feugiat, pretium velit eu, volutpat arcu. In luctus lorem et nunc venenatis, sed suscipit augue congue. Duis gravida scelerisque venenatis. Duis vehicula ac massa in pulvinar. Cras urna risus, dapibus a pellentesque a, convallis ut urna. Nunc at urna blandit, facilisis ex suscipit, condimentum massa. Phasellus auctor lacus velit, sit amet interdum augue tempus ut. Duis ante mauris, elementum vel commodo a, laoreet ac magna.

Etiam tellus lacus, rhoncus et leo ac, rhoncus vehicula elit. Vivamus condimentum sit amet ante eu interdum. Pellentesque facilisis ultricies orci, luctus sodales turpis rutrum in. Nullam felis enim, hendrerit ac risus quis, egestas tempor urna. Sed interdum nisl in elit malesuada vestibulum. Proin mattis interdum gravida. Aenean vitae justo nisi. Quisque nec ultrices velit.

Vestibulum at ligula nunc. Ut maximus et orci et eleifend. Nulla nec enim in tellus condimentum porttitor eget sed felis. Morbi a metus sit amet lorem varius volutpat porttitor ut orci. Donec varius ornare orci, vel feugiat quam egestas non. Etiam sed cursus ante. In sit amet dolor nunc. Maecenas tincidunt ac nulla in semper. Sed sit amet ultricies metus, et auctor erat. Morbi lacinia sem est, sit amet gravida ante mattis eu. Phasellus sit amet tristique elit. Praesent non finibus quam. Vivamus finibus tempus libero, at pretium orci commodo ut. Praesent semper lectus non massa pulvinar efficitur.

Donec pulvinar laoreet nulla. Sed rutrum egestas vehicula. Vivamus finibus consequat purus, et dignissim neque venenatis vitae. Duis vestibulum facilisis leo nec tincidunt. Fusce auctor quam sed ex blandit cursus. Aenean non commodo sem. Ut tincidunt felis at erat aliquam condimentum. Sed eros neque, sollicitudin non quam sit amet, gravida placerat velit. Mauris lobortis a erat ac porta. Vivamus nec gravida tortor, ac cursus nibh.

Mauris tincidunt mattis elementum. Aenean convallis urna sit amet cursus semper. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Aliquam fermentum in magna ac convallis. Morbi auctor purus in tincidunt consectetur. Aliquam purus justo, aliquam ut nunc dignissim, cursus imperdiet felis. Aenean quis dui at augue efficitur interdum in eu ligula. Quisque mi nulla, euismod sed ultrices sed, finibus feugiat risus. Duis in tempor nisi. In eu facilisis erat. Nullam aliquet, nunc eu dignissim faucibus, tortor odio blandit quam, id vulputate leo lectus vel neque. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae;

Nam nibh mauris, finibus a lobortis sit amet, elementum vitae dolor. Mauris eleifend condimentum tellus at mollis. Pellentesque eget pellentesque nisl, at malesuada odio. Aenean efficitur iaculis justo, at viverra velit volutpat non. Suspendisse eget diam malesuada, fringilla sapien a, mattis erat. Nulla facilisi. Cras feugiat massa a convallis auctor. Donec at tellus eu nisl congue feugiat in a quam. Ut et dictum nibh. Etiam purus enim, accumsan sollicitudin auctor a, volutpat at neque. Etiam aliquet tincidunt sollicitudin. Nam facilisis sapien sit amet justo luctus, non euismod mauris accumsan. Nam sem magna, tempus eu lacinia et, porta quis ante. Praesent auctor placerat viverra. Suspendisse eu pretium tortor.

Mauris pellentesque mi elit, vitae ultrices nulla pellentesque eget. Sed at arcu eu neque semper consequat non sit amet urna. Fusce aliquet facilisis ex, et vulputate nunc fringilla ut. Praesent dapibus, orci sed tempor auctor, nisl leo varius mi, eu pretium neque ipsum et nisi. Suspendisse potenti. Maecenas tincidunt et lorem a luctus. Phasellus gravida justo vitae ornare egestas.

Aenean tincidunt interdum neque id tristique. Duis placerat, augue eget convallis hendrerit, elit quam suscipit mi, id suscipit orci ligula nec urna. Aenean at arcu in tortor ullamcorper dictum. Sed mollis ex nunc, et consectetur arcu cursus vitae. Etiam blandit pharetra semper. Suspendisse tempor, tellus nec pellentesque ullamcorper, tellus dolor convallis nisl, ac aliquet nisi ante ut lacus. Aenean congue dolor blandit tellus pellentesque tempus. Nulla arcu purus, commodo luctus metus sed, aliquet feugiat quam. Sed volutpat bibendum eros, quis tristique nisl scelerisque vel.

Aliquam nec consequat erat, sed lacinia dolor. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nullam dignissim porttitor malesuada. Integer fringilla sapien lectus. Morbi laoreet luctus libero, eget commodo nisi sagittis id. Suspendisse sed orci enim. Nulla pharetra aliquet rutrum. Donec neque augue, ornare ut quam et, congue dignissim dolor. Fusce egestas eu magna nec ullamcorper. Curabitur eu sapien ac purus scelerisque vestibulum in a eros. Phasellus nec posuere lectus. Mauris mauris enim, placerat finibus leo ut, eleifend rhoncus ipsum. In sollicitudin commodo leo sit amet suscipit. Aenean sodales cursus ex, ac porttitor felis sagittis et. Quisque in scelerisque neque, at facilisis sapien. Etiam luctus, enim quis pulvinar tempus, nulla lectus malesuada mauris, pellentesque ornare felis dui semper dui.

Aliquam et congue enim, ut pretium lacus. Morbi pharetra in ipsum eu viverra. Vivamus sit amet sem quis libero semper egestas. In imperdiet molestie porta. Maecenas scelerisque nisl id enim condimentum, sed vehicula orci imperdiet. Suspendisse potenti. Vestibulum scelerisque dapibus mi, quis aliquam tellus. Cras tellus purus, vulputate at neque quis, lacinia luctus mi.

Quisque ex nisi, eleifend eget varius sed, luctus eget lorem. Donec vel arcu at nulla placerat bibendum at at libero. Nam id iaculis elit, sit amet varius ex. Cras ornare, purus nec bibendum scelerisque, purus mi sagittis mi, in mattis tortor lorem ut augue. Nam facilisis vehicula facilisis. Aliquam commodo porttitor fermentum. Vestibulum quis imperdiet orci, non pretium lacus. Cras eleifend at arcu sit amet commodo. Vivamus pretium tellus convallis justo pretium facilisis. Maecenas quis est sed mi fermentum accumsan. Suspendisse risus est, pretium eu metus vel, dictum iaculis ligula.

Morbi elementum tincidunt nunc eu imperdiet. Nam convallis maximus metus a facilisis. Praesent luctus felis vitae orci aliquam posuere. Nulla posuere fringilla eros, ut vehicula dui faucibus id. Nulla ut viverra augue. Fusce vitae mi non quam sagittis porttitor at rutrum libero. Nulla facilisi. Nunc semper sem eget ligula luctus, nec gravida massa consectetur. Pellentesque interdum viverra velit a varius. Praesent pretium libero nibh, pharetra consectetur lacus viverra in.

Mauris a facilisis sapien, a porttitor velit. Nulla vehicula ipsum ipsum, vel maximus quam gravida sed. Nullam nibh mauris, commodo auctor eros id, venenatis accumsan ex. Cras lobortis magna at orci ultricies sollicitudin. In non augue fermentum, vulputate turpis commodo, condimentum lorem. Cras augue est, fermentum sit amet vulputate at, condimentum in purus. Vivamus vestibulum nunc leo, a ultricies tellus congue nec.

Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed posuere sapien a urna scelerisque, sed rutrum velit efficitur. Nam accumsan, velit vitae dapibus ullamcorper, nisi dolor iaculis magna, vel rhoncus purus elit eget felis. Duis venenatis justo eu tellus mattis aliquam. Curabitur pharetra dui a erat varius sagittis. Nunc vulputate semper lectus, at lacinia leo rutrum ac. Vivamus vulputate tellus sit amet commodo commodo. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Etiam feugiat dui id magna hendrerit, sed dapibus leo suscipit. Vivamus sit amet rhoncus turpis. Integer ligula elit, facilisis nec facilisis elementum, molestie vitae erat. Proin sed magna ullamcorper, dignissim nulla id, ullamcorper urna. Nullam in eros hendrerit sapien luctus posuere. Ut elit dolor, lacinia at pharetra nec, rutrum at est.

Sed commodo posuere metus, id porttitor dui tempus vitae. Pellentesque at malesuada libero. Donec in tristique odio. Etiam vehicula interdum aliquam. Nunc vel erat sem. Donec finibus tortor vel sem pretium pulvinar. Vestibulum auctor elit ut lorem tempor, non consectetur purus tincidunt. Maecenas lacinia nulla vitae ligula volutpat vestibulum. Donec rhoncus quam nec enim scelerisque, sed molestie dolor semper. Curabitur id elementum augue. Ut lobortis ante sodales turpis ultrices, et condimentum dolor condimentum.

In bibendum facilisis nunc, eget pellentesque purus varius vitae. Morbi et magna augue. Donec viverra est vitae nibh consequat euismod. Fusce sed odio vehicula, malesuada mi eu, volutpat nunc. Phasellus mattis massa sit amet purus placerat, sed laoreet risus volutpat. Etiam suscipit interdum enim, in faucibus enim malesuada eu. In erat sem, ornare in purus vitae, sodales laoreet lacus. Duis sollicitudin sed augue sit amet imperdiet. Nam in lobortis est. Donec ultricies diam id convallis blandit. Nam sed nibh neque.

In fringilla eros et justo aliquam dapibus. Vivamus et porttitor orci. Donec sit amet elit eget nunc tempus consectetur. Mauris enim est, congue vitae interdum eget, molestie sed nunc. Proin non tellus sem. Nam ut scelerisque nisl, ut placerat neque. Vestibulum ut lectus sed neque vulputate vestibulum ac non sapien. Suspendisse ornare orci est, eget vehicula mi porta et. Cras eu vehicula erat, id facilisis tortor. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Phasellus id tempus elit, sed viverra metus. In sed convallis risus. Proin rutrum accumsan elementum. Aliquam non posuere ante, nec molestie mi.

Vestibulum volutpat consectetur lacus, ut ullamcorper ligula sodales ac. Sed ullamcorper aliquam congue. Nullam at lacus finibus, mollis orci quis, feugiat massa. Donec sed dui et mi ullamcorper maximus ac eu massa. In hac habitasse platea dictumst. Cras magna eros, tempor finibus lectus at, congue aliquet libero. Suspendisse eleifend tortor id tortor blandit, in vestibulum leo lobortis. Aenean mollis turpis at odio accumsan congue. Aliquam quis ante ut metus mollis porttitor. Donec a tellus eget nunc iaculis laoreet quis vulputate velit. Duis non nunc fringilla, ultrices elit nec, imperdiet lectus. Maecenas eu sapien sagittis, pharetra arcu sit amet, tristique purus. Ut luctus commodo eros, et gravida nibh auctor quis.

Praesent nec augue ut lectus porta vulputate. Integer bibendum malesuada suscipit. Suspendisse sit amet odio ac turpis efficitur ultricies sed id magna. Vivamus porta sed magna ut rutrum. Phasellus sit amet arcu nec ante faucibus condimentum a tempus nunc. Curabitur nec venenatis purus. Maecenas at gravida urna, sed aliquam ante. Duis eget metus viverra, consequat justo eu, viverra urna. Ut placerat odio lorem, sed egestas metus tincidunt vitae. Sed bibendum accumsan maximus.

Nullam mollis interdum nulla at elementum. Morbi ac finibus nisi. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce nec sem metus. Aliquam efficitur lorem quis nunc lobortis, vitae tristique enim vulputate. Praesent feugiat elit eget blandit condimentum. Vivamus bibendum id ipsum sit amet semper. Vestibulum nec facilisis orci. Duis suscipit elit tortor, nec porttitor arcu elementum nec. Vestibulum fringilla condimentum egestas. Etiam lacus purus, varius vitae augue sit amet, pharetra viverra lorem. Proin turpis justo, efficitur tincidunt dignissim eu, ornare vitae ex. Aliquam erat volutpat. Donec et blandit quam. Mauris tristique sapien sapien, non auctor ligula interdum vestibulum. Maecenas iaculis nunc sem, et interdum magna iaculis lobortis.

Donec ante tortor, hendrerit quis facilisis sed, eleifend sed diam. Duis metus felis, fermentum ut ornare eu, rhoncus posuere risus. Quisque euismod auctor luctus. Pellentesque ac tellus eget dui aliquet condimentum eu vel tellus. Etiam vitae nisi porttitor, feugiat nunc quis, vehicula metus. Morbi a nisi id nisi varius dapibus. Curabitur quis nunc elit. Integer consectetur risus sed sem varius dignissim. Mauris suscipit at nisi ut vulputate.

Etiam ultricies risus a facilisis sollicitudin. Aenean vitae sapien sit amet orci interdum placerat. Sed et metus et justo venenatis efficitur eu sed mauris. Curabitur porta sem est, sit amet placerat libero iaculis vitae. Phasellus porta, augue nec vehicula condimentum, risus ipsum pulvinar erat, vitae facilisis eros erat consectetur enim. Phasellus hendrerit luctus tortor, nec varius magna rutrum in. Suspendisse pretium ipsum eu nulla ultricies tincidunt. Cras commodo facilisis mi, eu pretium nulla ornare sit amet. Nullam suscipit est felis, pellentesque mollis est auctor sit amet. Fusce fringilla commodo semper. Fusce tristique euismod tincidunt. Integer vel tempus sem, ac egestas eros. Cras posuere ipsum eget sapien congue, ullamcorper lobortis dolor consequat. Curabitur malesuada sapien quis lorem ultrices dapibus. Integer semper, ipsum hendrerit lacinia ornare, tellus lacus aliquet libero, quis ornare ligula quam nec erat. Maecenas arcu turpis, sollicitudin ac odio in, elementum feugiat orci.

In commodo lorem eu sodales rutrum. Curabitur fringilla aliquam libero, at pellentesque eros maximus non. Morbi ut ligula ut ipsum vehicula malesuada eu eu arcu. Fusce dignissim ex quis libero porttitor sodales. Nam eget erat magna. Suspendisse vitae posuere neque. Nulla at tincidunt nibh. Maecenas consectetur eget elit a molestie. Etiam eleifend urna purus, non condimentum tellus porta vel. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nullam molestie sit amet urna ut porta. Fusce porttitor bibendum risus, fermentum iaculis metus pretium ac. Vestibulum rhoncus nulla turpis, at viverra risus vulputate id. Duis vel neque vitae justo hendrerit ultrices sit amet a neque.

Praesent et enim id purus dictum laoreet ac in est. Vivamus placerat libero sed diam mollis, id lobortis ligula vestibulum. Suspendisse id eros ante. Fusce sed diam faucibus, vestibulum massa eget, aliquam enim. Cras convallis dui libero, at vestibulum lacus congue quis. Vestibulum rutrum dolor ut diam congue euismod. Donec eros nulla, vestibulum eu dignissim id, finibus vel nibh.

Praesent nec lacus euismod, placerat erat vitae, ultricies metus. Sed viverra nibh a cursus tincidunt. Ut tincidunt sed mauris id fermentum. Sed orci mi, porta quis tempus eu, pretium non risus. Nunc id leo ornare velit consectetur sodales pharetra et justo. Donec at purus vel velit aliquam commodo. Aenean pellentesque fringilla tincidunt. Sed tempor pellentesque urna, sed pharetra lacus luctus a. Ut dui tellus, euismod nec iaculis sit amet, ultricies at augue. Ut sagittis turpis ligula, a maximus felis vestibulum non. Ut ac dapibus velit, et bibendum augue. Proin quis posuere tellus. Maecenas euismod, ipsum ac dictum commodo, tortor nisi sagittis dolor, at pretium orci ante vel mauris. Ut convallis et nisi id posuere. Mauris bibendum mollis nisi eget viverra.

Integer nec eros vel felis posuere molestie. Sed nec lacus risus. Ut massa magna, maximus et est vitae, tempus ultricies metus. Nam et diam ipsum. Aliquam interdum eu felis vitae consequat. Duis elit ante, consequat in nulla at, posuere tempus justo. Integer tincidunt quam a ligula facilisis, ut condimentum mauris commodo. Suspendisse vel lorem tincidunt, hendrerit ex nec, mattis purus. Pellentesque at varius turpis, a feugiat augue. Praesent hendrerit nulla a libero tempor molestie.

Nam ultrices auctor nibh a vehicula. Aenean sodales nec lacus sit amet auctor. Nulla vitae justo a lacus placerat suscipit molestie in odio. Aliquam erat volutpat. Cras a risus dapibus, convallis ante ac, euismod mi. Duis pellentesque metus nibh, eget pulvinar nulla mollis eu. Vestibulum vestibulum nunc scelerisque vehicula molestie. Quisque scelerisque arcu non turpis porta faucibus. Aenean ac nisl tempus ligula mollis sagittis.

Curabitur sollicitudin viverra suscipit. Mauris ac viverra nulla, vel sollicitudin nulla. Sed eget massa ullamcorper, pellentesque arcu in, auctor neque. Proin imperdiet sagittis felis sit amet malesuada. Etiam hendrerit accumsan finibus. Cras dignissim tincidunt vestibulum. Aliquam accumsan quam quam, sit amet tempor eros eleifend eget. Nulla id scelerisque est. Cras nec ante at nisl imperdiet porta a id nulla. Aliquam turpis urna, condimentum at rutrum nec, rutrum malesuada dolor. Maecenas pharetra, velit ut pharetra accumsan, enim metus vehicula dui, id venenatis sem metus sed odio. Maecenas aliquet et orci vitae aliquet. Interdum et malesuada fames ac ante ipsum primis in faucibus. Aliquam erat volutpat. Phasellus iaculis consectetur consectetur.

Integer euismod posuere laoreet. Nunc dapibus, lacus et facilisis rutrum, purus sapien mollis dolor, vel maximus augue tortor at eros. Phasellus sit amet urna ut velit varius auctor sit amet imperdiet diam. Nulla in velit nunc. Quisque tempor sem fringilla accumsan sagittis. Cras accumsan tristique fringilla. Vestibulum arcu lacus, tristique ac laoreet non, pulvinar ac lorem. Duis id lacus sapien. Sed finibus massa id mi ultricies laoreet. Maecenas efficitur dui sed consequat dignissim.

Praesent convallis, sem ut posuere tempor, velit massa molestie elit, ac sollicitudin leo nibh eu nibh. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Praesent cursus velit ut nulla commodo, nec interdum nisi lobortis. Nulla mi velit, condimentum at eleifend sit amet, maximus in neque. Nulla vel mi nulla. Aliquam eleifend nec sem vitae tincidunt. In suscipit placerat nulla, id efficitur purus bibendum vel. Suspendisse id pretium elit. Vestibulum et dolor massa. Duis sagittis magna velit, at euismod metus faucibus ut. Nulla sagittis facilisis nisl, imperdiet ornare nunc vestibulum sed. Praesent blandit laoreet eros sed tristique. Aliquam ut aliquam augue, eu suscipit nunc. Aliquam consequat aliquet justo id scelerisque.

Nulla turpis sapien, vestibulum tempor dictum placerat, posuere eget nibh. Morbi vehicula lectus eu tortor elementum fringilla. Fusce sit amet lacinia nisl. Vivamus id metus justo. Mauris sed suscipit ex. Pellentesque imperdiet, dolor at lacinia luctus, eros dui sagittis enim, ac vehicula augue enim non sem. Fusce porttitor risus id condimentum rhoncus. Aenean ut tincidunt est, a aliquet nulla. Donec molestie sem at suscipit faucibus. Mauris laoreet felis odio, et tristique ante rutrum vel.

Phasellus vitae mauris consectetur, ullamcorper sem eu, condimentum magna. Sed rhoncus aliquam feugiat. Quisque lacinia ante eget dui pharetra, a maximus massa pretium. Interdum et malesuada fames ac ante ipsum primis in faucibus. Aenean consequat pharetra cursus. Curabitur ac odio efficitur felis fringilla posuere a vel lacus. Vestibulum hendrerit felis tellus, quis luctus augue condimentum ut. Nulla euismod lectus eget cursus sodales. Pellentesque aliquet leo id ipsum eleifend placerat.

Etiam porta scelerisque dui, nec efficitur tortor dictum in. Mauris leo quam, pulvinar sit amet nisi a, cursus bibendum justo. Aenean leo erat, ornare nec mauris sed, ultricies tincidunt ligula. Duis dictum in mi sit amet aliquam. In hac habitasse platea dictumst. In tempor ligula sapien, sed tristique nulla varius quis. Morbi eget lorem eros.

Proin quis mauris laoreet, varius nunc et, euismod lorem. Phasellus sit amet interdum est, vitae elementum sem. Etiam consectetur orci vel rhoncus fringilla. Mauris laoreet suscipit risus eget semper. Mauris elementum, velit eget volutpat cursus, lacus sapien dictum est, vulputate auctor ex turpis quis felis. Donec diam nulla, efficitur et efficitur eu, porta sed augue. Mauris malesuada pellentesque blandit. Cras molestie enim erat, sed consectetur sapien varius a.

Curabitur diam nisi, fringilla sed placerat ac, dignissim sit amet tortor. Nullam id tempus arcu, ac eleifend orci. Cras tempus enim non metus viverra finibus. Ut vitae diam in mi pellentesque semper. Aenean vestibulum, arcu ac fringilla interdum, augue dui porttitor lacus, id lacinia sem arcu quis lectus. Aliquam vestibulum ac odio eu feugiat. Integer semper eleifend nibh, vestibulum maximus erat ultricies vel.

Pellentesque finibus vel purus et consequat. Duis at quam sodales, rutrum justo vitae, volutpat ante. Curabitur tempor lacinia dolor, sed convallis lacus porta et. Pellentesque tempus tempus fringilla. Nullam porttitor gravida nunc, nec eleifend dolor tempus vel. Suspendisse mollis sem vel libero facilisis pellentesque quis vitae lectus. Ut porttitor, erat eget tincidunt bibendum, magna quam posuere nulla, in scelerisque est ante non eros. Vivamus scelerisque tellus sit amet pretium rutrum.

Donec nec arcu at ipsum ultrices vulputate eget a urna. Vivamus at interdum quam. Curabitur faucibus tempus feugiat. Vestibulum iaculis tempus ipsum, a dapibus orci malesuada et. Sed finibus laoreet felis, a condimentum ipsum facilisis nec. Etiam gravida at risus non rhoncus. Vivamus laoreet aliquet molestie. Curabitur convallis ac turpis id varius. Nunc mollis justo vel velit semper, et pellentesque turpis hendrerit.

Fusce accumsan tortor elit, sed dapibus orci pharetra quis. Etiam rhoncus dolor arcu, id semper nunc sollicitudin nec. Donec blandit ornare nisl, in varius magna fermentum vel. Etiam nec mattis elit. Vestibulum vitae sagittis sem. Nulla a iaculis ligula. Mauris pharetra, neque sit amet varius pretium, diam velit tincidunt mi, non pellentesque sem lorem nec dui. Nullam lacinia tortor erat. Integer aliquet tincidunt ligula. Sed commodo laoreet consectetur.

Morbi hendrerit blandit sodales. Phasellus tristique feugiat ex, sed ornare massa molestie id. Donec nec aliquam tellus, ut mattis turpis. Aenean porttitor ante in lorem malesuada dignissim. Mauris placerat accumsan risus, non ultrices mi hendrerit at. Phasellus vestibulum felis et est tincidunt elementum. Nam at quam sapien.

Sed a interdum neque. Nulla facilisi. Donec urna libero, scelerisque at turpis a, semper sagittis nibh. Suspendisse lorem metus, fringilla eu tortor vitae, mattis ultricies felis. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Nulla id imperdiet leo, id pellentesque risus. Interdum et malesuada fames ac ante ipsum primis in faucibus.

Suspendisse a velit ut mi eleifend aliquet sed a libero. Vivamus a eleifend mauris, vel consectetur ipsum. Ut mi odio, scelerisque vitae aliquam quis, scelerisque bibendum nulla. Morbi cursus ligula ut tempor finibus. Praesent bibendum nisl sed elementum suscipit. Etiam aliquam sit amet leo vitae congue. Nunc commodo laoreet enim, at semper ante fermentum eget. Donec ultrices tempus nisl in ornare. Sed mollis justo ut ipsum tempor, vel pellentesque nulla tempus. Nam auctor vestibulum leo bibendum gravida. Quisque cursus, libero sit amet tincidunt tincidunt, leo nisl ultrices justo, ac posuere felis odio sit amet nunc. Mauris non iaculis justo. In erat eros, luctus quis lobortis et, tincidunt consequat purus.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent viverra, urna vel tempus molestie, massa ante pulvinar urna, eget finibus nunc ligula non mauris. Interdum et malesuada fames ac ante ipsum primis in faucibus. In euismod mollis auctor. Cras maximus libero at lacus auctor egestas. Etiam sed odio vitae velit consequat vehicula. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Vivamus venenatis ex a ipsum sodales, nec placerat orci congue. Etiam bibendum ante elit, sit amet bibendum arcu convallis vitae. Mauris in hendrerit massa. Maecenas eu dolor eget ipsum fermentum commodo at sagittis ante. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Mauris et ante congue, suscipit nisi id, ullamcorper lorem. Donec ac lectus leo. Pellentesque pharetra ex tristique, imperdiet turpis porta, maximus leo.

Suspendisse scelerisque dui at enim tincidunt mollis. Integer urna magna, placerat quis lectus quis, faucibus blandit dolor. Integer vulputate, dui sit amet lacinia mollis, nibh velit placerat eros, a lacinia ligula dolor at nibh. Ut sagittis leo libero, ut iaculis orci volutpat ut. Cras dictum ipsum rutrum vehicula ullamcorper. Integer efficitur, mauris id pulvinar tristique, leo dui mollis diam, efficitur volutpat nisl mi vel lectus. Fusce eu lectus enim.

Duis hendrerit pharetra posuere. Proin ac ipsum quis lorem feugiat maximus vel in ex. Mauris feugiat sodales massa, pharetra mattis nibh faucibus eu. Nunc hendrerit cursus aliquam. Donec pellentesque, sem at finibus ullamcorper, nibh tortor vulputate leo, at finibus odio erat vitae nisi. Curabitur ornare blandit ultrices. In porta nunc quis purus egestas malesuada. Nullam vestibulum velit sit amet erat ultricies, in rutrum ex finibus. Aenean id consequat metus. Cras dapibus feugiat scelerisque. Vestibulum ligula turpis, suscipit quis augue non, cursus auctor diam.

Sed mollis luctus metus, id accumsan sem efficitur non. Praesent a accumsan libero, ac tempus mauris. Sed cursus dapibus sollicitudin. Etiam at mattis ex, id ultricies quam. Donec vulputate placerat diam at accumsan. Aliquam ac lectus urna. Aliquam bibendum tincidunt turpis sed dignissim. Phasellus arcu lectus, lacinia eu tortor nec, consequat tincidunt eros. Integer ac lorem mauris. Nulla sed sollicitudin risus. Integer eget iaculis ante, elementum scelerisque est. Sed eget dapibus metus, eu euismod orci. Mauris feugiat facilisis gravida. Pellentesque convallis ex metus, ac faucibus orci laoreet ac. Maecenas volutpat nulla ut odio imperdiet hendrerit.

Aliquam ac lacinia nibh. Quisque lectus tellus, eleifend a orci eu, fringilla rhoncus leo. Sed tellus ipsum, interdum a libero nec, venenatis porttitor leo. Integer vel placerat sapien. Praesent viverra ultrices magna quis pulvinar. Curabitur vulputate gravida ligula et suscipit. Nullam accumsan dui et mollis commodo. Curabitur id elementum nunc. Donec ut vestibulum augue, vitae suscipit massa. Donec fringilla pharetra sapien, id rhoncus ex eleifend eget. Pellentesque ultrices sed lectus sed iaculis. Quisque vitae magna quis diam ultricies tincidunt. Ut condimentum tellus id dui pulvinar pretium.

Aenean vehicula auctor mauris, quis imperdiet elit feugiat id. Duis efficitur lectus sed tellus egestas congue. Curabitur ullamcorper sem erat, in lacinia neque semper eu. Nunc in bibendum elit. Curabitur vulputate turpis vitae lorem vehicula pharetra. Aenean nibh turpis, malesuada id ligula quis, bibendum hendrerit nisl. Aenean et odio laoreet, eleifend nibh in, consectetur erat. Nullam bibendum diam nisl, in mattis erat vestibulum ultrices. Suspendisse lobortis, ex et tristique malesuada, est lectus sagittis metus, quis condimentum lacus metus vitae libero. Etiam dapibus efficitur augue et vulputate. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vivamus et molestie nunc, sit amet iaculis quam. Nam rutrum felis nec pulvinar accumsan. Nullam ac turpis congue, pulvinar turpis in, lacinia nunc. Aliquam ullamcorper rutrum ante, a dignissim nunc lacinia nec. Donec vel tempus quam, mollis tincidunt felis.

Aenean facilisis a enim nec molestie. Vivamus in orci vitae neque consectetur lacinia. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed viverra, diam vitae hendrerit placerat, tortor ante mollis mauris, nec gravida nibh massa nec leo. Fusce fermentum nunc odio, id dapibus felis eleifend eget. Suspendisse potenti. Vestibulum tristique massa turpis, fermentum mattis nisi cursus at. Morbi lobortis nisi eu tellus euismod, vitae efficitur mauris egestas. Sed nec elit ac arcu porta vestibulum eu sit amet arcu.

Vivamus nulla magna, maximus sed placerat ut, vestibulum vitae ipsum. Suspendisse potenti. Praesent lobortis nisl porta, sollicitudin purus posuere, posuere mi. Sed tempor velit non erat maximus mattis. Sed venenatis sit amet neque nec bibendum. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Vestibulum nulla ex, lacinia in sagittis eget, tempor eu nibh. Integer porttitor lorem id malesuada lacinia. Sed est massa, congue at aliquet sed, dictum et tellus. Vestibulum vitae rhoncus nulla. Praesent diam sem, viverra quis nibh sit amet, commodo faucibus sem. Nullam rutrum rhoncus turpis, eu scelerisque lectus malesuada vitae. Praesent fringilla velit quis suscipit egestas. Mauris efficitur ligula id elit consequat luctus. Sed a dui vitae justo egestas ultrices. Morbi posuere, ipsum at semper eleifend, est mauris faucibus eros, ut vehicula est nunc nec diam.

Vivamus aliquet vehicula tortor, et pharetra magna congue vitae. Donec vitae erat facilisis orci viverra pharetra sed id arcu. Suspendisse potenti. Etiam nibh nisl, ornare id rutrum a, suscipit ac nisi. Mauris eu congue lorem, eu sagittis ligula. Duis sed felis consectetur, porttitor libero eu, lobortis libero. Donec finibus ornare velit. Vestibulum dignissim sit amet justo id commodo. Sed fermentum sollicitudin libero. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Pellentesque elementum ex sit amet odio condimentum feugiat. Pellentesque maximus fringilla orci ut pharetra. Quisque congue quis lacus et interdum.

Aliquam id leo lorem. Phasellus vestibulum, nunc sed laoreet tincidunt, nisi neque feugiat felis, eget fringilla augue purus a ipsum. Aliquam id imperdiet mi, non rhoncus ex. Etiam accumsan nisi nec diam facilisis condimentum. Mauris molestie massa nec nunc porta dignissim. Donec ut finibus ante, id fringilla dui. Etiam lobortis sodales purus, eu tincidunt ligula egestas non. Pellentesque eget elit erat. Maecenas dapibus consequat nibh. Proin in arcu arcu. Pellentesque dignissim dignissim arcu vel dapibus.

In efficitur ante lectus, id facilisis mi vestibulum et. Sed bibendum elementum magna, ac varius risus rutrum malesuada. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nam quis placerat tellus, eu iaculis libero. Etiam ut suscipit dui. Suspendisse suscipit aliquam urna, rutrum laoreet augue scelerisque non. Suspendisse efficitur eu tortor vel vulputate. Mauris lacus nisl, hendrerit ut tellus eget, tincidunt fermentum felis. Sed sit amet dolor et turpis mattis aliquam convallis suscipit massa. Nunc vel nulla quam. Ut condimentum lorem vehicula, ullamcorper massa nec, tincidunt libero. Vestibulum scelerisque, risus sagittis ultricies sodales, ex elit vestibulum orci, a laoreet lorem arcu in tellus. Aliquam tristique consequat nisl ac luctus. Sed finibus aliquet urna vel tincidunt. Nulla eget facilisis augue.

Mauris fermentum neque urna, ut blandit ligula faucibus non. Nunc sit amet sapien arcu. Vestibulum risus erat, dignissim a tortor id, eleifend cursus leo. Morbi finibus elementum orci vel congue. Donec auctor eleifend bibendum. Curabitur dapibus felis non congue aliquam. Duis placerat neque non urna venenatis sagittis. Praesent lectus elit, efficitur a rhoncus sit amet, commodo id quam. Vivamus aliquet elit sed fringilla feugiat. Phasellus in felis fringilla turpis iaculis pretium sed eget dui. Nulla vehicula, lectus vel pretium aliquet, purus risus tempus turpis, in semper tortor orci at tellus. Nam malesuada purus nibh, et volutpat elit consectetur eu. Phasellus pulvinar, magna et viverra imperdiet, elit sem fringilla sem, id feugiat mauris lectus ac nulla. Integer porta risus vitae lacus lobortis pretium.

Morbi viverra at arcu in faucibus. Nullam elit augue, mattis quis fermentum finibus, ultricies at lacus. Aliquam nec sapien lectus. In ex leo, cursus et imperdiet vel, tempus vel nisi. Ut accumsan lectus aliquet ullamcorper mattis. Nunc iaculis elit semper quam dictum, quis maximus mi consectetur. Integer non molestie dui. Suspendisse lacinia orci sapien, in porta neque consectetur id. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce feugiat auctor purus, auctor accumsan mi porttitor quis. Duis ac odio urna. Nam maximus fringilla lorem, non interdum nunc. Proin fermentum neque eu commodo vehicula. Fusce odio mi, mattis sit amet vestibulum id, bibendum in ipsum. Quisque a faucibus mauris, in pretium nibh. Sed lobortis, est sed rhoncus pellentesque, arcu lorem efficitur elit, sed tincidunt turpis justo a turpis.

Suspendisse sit amet est efficitur lorem aliquam tincidunt. In hendrerit ligula nec risus egestas pulvinar. Nunc at consequat leo, at dictum metus. Morbi tristique felis et velit commodo, non egestas sapien pharetra. Donec elit turpis, malesuada eget tellus eget, sodales varius felis. In risus diam, mattis non elementum non, cursus vel ligula. Quisque rhoncus, justo ac pulvinar vehicula, tellus mauris egestas tellus, eget imperdiet tellus arcu nullam.`

	go func() {
		fmt.Fprintf(secureCW, message)
		upW.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf, err := ioutil.ReadAll(secureCR)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	expected := message
	if got := string(buf); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
	}
}
