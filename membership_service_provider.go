package fabclient

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

type membershipServiceProvider interface {
	createSigningIdentity(certificate, privateKey string) (mspprovider.SigningIdentity, error)
	getSigningIdentity(id string) (mspprovider.SigningIdentity, error)
}

type membershipServiceClient struct {
	client *msp.Client
}

func newMembershipServiceProvider(ctx context.ClientProvider, organization string) (membershipServiceProvider, error) {
	client, err := msp.New(ctx, msp.WithOrg(organization))
	if err != nil {
		return nil, err
	}

	mspclient := &membershipServiceClient{
		client: client,
	}

	return mspclient, nil
}

var _ membershipServiceProvider = (*membershipServiceClient)(nil)

func (m *membershipServiceClient) createSigningIdentity(certificate, privateKey string) (mspprovider.SigningIdentity, error) {
	certificateAsBytes, err := ioutil.ReadFile(certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	privateKeyAsBytes, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	return m.client.CreateSigningIdentity(mspprovider.WithCert(certificateAsBytes), mspprovider.WithPrivateKey(privateKeyAsBytes))
}

func (m *membershipServiceClient) getSigningIdentity(id string) (mspprovider.SigningIdentity, error) {
	return m.client.GetSigningIdentity(id)
}
