#ifndef CONNECTION_H
#define CONNECTION_H
#include <QObject>
#include <QTcpSocket>
#include <QTimer>

// ConnectionThread是一次性的, 用过一次应当清理, 禁止重用
class ConnectionThread: public QObject {
	Q_OBJECT

private slots:
	/*controller控制线程*/
	void onStartConnectionEmited(QString, uint);
	void onCloseConnectionEmited();
	void onMessageSent(QString);

	/*socket的槽*/
	void onSocketErrorOccured(QAbstractSocket::SocketError);
	void onSocketReadyRead();
	void onSocketConnected();
	void onSocketDisconnected();

	/*timer的槽*/
	void onTimerShot();

signals:
	/*通知controller*/
	void messageArrived(QString);
	// errorOccured报告包括心跳包3次超时的错误, 和socket的error
	void errorOccured(QString);
	void socketConnected();
	void socketDisconnected();

public:
	ConnectionThread(QObject*);

private:
	QObject *controller;
	QTcpSocket *socket;
	QTimer *timer = nullptr;
	int timeoutCount = 0;

	static QByteArray heartPackBytes;
	static int timeout;
	static int timeoutMax;

	static uint32_t toUint32(char *b) {
		return uint32_t(b[0]) | uint32_t(b[1]) << 8 | uint32_t(b[2]) << 16 | uint32_t(b[3]) << 24;
	}

	static void fromUint32(char *b, uint32_t v) {
		b[0] = char(v);
		b[1] = char(v >> 8);
		b[2] = char(v >> 16);
		b[3] = char(v >> 24);
	}
};

#endif // CONNECTION_H
