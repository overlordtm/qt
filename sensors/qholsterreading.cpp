#include "qholsterreading.h"
#include <QString>
#include <QVariant>
#include <QUrl>
#include <QModelIndex>
#include <QHolsterReading>
#include "_cgo_export.h"

class MyQHolsterReading: public QHolsterReading {
public:
};

int QHolsterReading_Holstered(void* ptr){
	return static_cast<QHolsterReading*>(ptr)->holstered();
}

void QHolsterReading_SetHolstered(void* ptr, int holstered){
	static_cast<QHolsterReading*>(ptr)->setHolstered(holstered != 0);
}
