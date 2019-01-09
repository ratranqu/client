// @flow
import * as React from 'react'
import {action, storiesOf, unexpected} from '../../../../stories/storybook'
import {Box} from '../../../../common-adapters'
import {globalColors} from '../../../../styles'
import Payment from '.'

const common = {
  approxWorth: '',
  cancelButtonInfo: '',
  cancelButtonLabel: '',
  canceled: false,
  claimButtonLabel: '',
  onCancel: action('onCancel'),
  onClaim: action('onClaim'),
  onSend: action('onSend'),
  sendButtonLabel: '',
}

const sentProps = {
  ...common,
  action: 'sent Lumens worth',
  amount: '$35',
  balanceChange: '-90.5700999 XLM',
  balanceChangeColor: globalColors.black_75,
  icon: null,
  loading: false,
  memo: ':beer:',
  pending: false,
}

const sentXLMProps = {
  ...common,
  action: 'sent',
  amount: '1 XLM',
  approxWorth: '$901.23 USD',
  balanceChange: '+1 XLM',
  balanceChangeColor: globalColors.green,
  icon: null,
  loading: false,
  memo: 'here you go',
  pending: false,
}

const sentNoMemoProps = {
  ...sentProps,
  memo: '',
}

const sendingProps = {
  ...common,
  action: 'sending Lumens worth',
  amount: '$35',
  balanceChange: '-90.5700999 XLM',
  balanceChangeColor: globalColors.grey,
  icon: 'iconfont-time',
  loading: false,
  memo: ':beer:',
  pending: true,
}

const claimableProps = {
  ...sendingProps,
  claimButtonLabel: 'Claim lumens worth',
}

const cancelableProps = {
  ...sendingProps,
  cancelButtonInfo: `This transaction can be canceled because barb does not yet have a wallet. Encourage barb to claim this and set up a wallet.`,
  cancelButtonLabel: 'Cancel',
}

const requestCommon = {
  ...common,
  action: 'requested Lumens worth',
  balanceChange: '',
  balanceChangeColor: '',
  icon: 'iconfont-stellar-request',
  loading: false,
  pending: false,
}

const youRequestProps = {
  ...requestCommon,
  amount: '$34',
  memo: 'for beers',
  onSend: unexpected('onSend'),
}

const theyRequestProps = {
  ...requestCommon,
  amount: '$107',
  memo: 'things',
  onSend: action('onSend'),
  sendButtonLabel: 'Send Lumens worth',
}

const sentAssetProps = {
  ...common,
  action: 'sent',
  amount: '1 BTC/Abc.def',
  balanceChange: '-1 BTC',
  balanceChangeColor: globalColors.black_75,
  icon: null,
  loading: false,
  memo: '₿',
  pending: false,
}

const loadingProps = {
  ...common,
  action: '',
  amount: '',
  balanceChange: '',
  balanceChangeColor: '',
  icon: null,
  loading: true,
  memo: '',
  pending: false,
}

const load = () => {
  storiesOf('Chat/Conversation/Account payments', module)
    .addDecorator(story => <Box style={{maxWidth: 420}}>{story()}</Box>)
    .add('Sent', () => <Payment {...sentProps} />)
    .add('Sent XLM', () => <Payment {...sentXLMProps} />)
    .add('Sent (no memo)', () => <Payment {...sentNoMemoProps} />)
    .add('Sending', () => <Payment {...sendingProps} />)
    .add(`Relay from sender's perspective`, () => <Payment {...cancelableProps} />)
    .add(`Relay from recipient's perspective`, () => <Payment {...claimableProps} />)
    .add('You request', () => <Payment {...youRequestProps} />)
    .add('They request', () => <Payment {...theyRequestProps} />)
    .add('Sent non-native', () => <Payment {...sentAssetProps} />)
    .add('Loading', () => <Payment {...loadingProps} />)
}

export default load
