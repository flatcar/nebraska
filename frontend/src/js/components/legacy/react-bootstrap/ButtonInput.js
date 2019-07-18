import PropTypes from 'prop-types';
import React from 'react';
import { FormGroup, Button } from 'react-bootstrap';
import InputBase from './InputBase.js';

class ButtonInput extends InputBase {
  renderFormGroup(children) {
    let {bsStyle, value, ...other} = this.props;
    return <FormGroup {...other}>{children}</FormGroup>;
  }

  renderInput() {
    let {value, ...other} = this.props;
    return <Button {...other} componentClass="input" ref="input" key="input" value={value} />;
  }
}

ButtonInput.types = Button.types;

ButtonInput.defaultProps = {
  type: 'button'
};

ButtonInput.propTypes = {
  type: PropTypes.oneOf(ButtonInput.types),
  bsStyle() {
    // defer to Button propTypes of bsStyle
    return null;
  },
  value: PropTypes.string
};

export default ButtonInput;
