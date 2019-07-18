import PropTypes from 'prop-types';
import React from 'react';
import Radio from './Radio';
import InputContainer from './InputContainer';
import classNames from 'classnames';

export default class RadioGroup extends InputContainer {
    constructor(props) {
        super(props);

        this.state = {
            value: props.defaultValue || props.value
        };
    }

    componentWillMount() {
        super.componentWillMount();

        this.props._registerInput(this);
    }

    componentWillUnmount() {
        super.componentWillUnmount();

        this.props._unregisterInput(this);
    }

    getValue() {
        let input = this._inputs[ this.props.name ];

        let value;

        input.forEach(ipt => {
            if (ipt.getChecked()) {
                value = ipt.getValue();
            }
        });

        return value;
    }

    render() {
        let label;

        if (this.props.label) {
            label = (
                <label className={classNames('control-label', this.props.labelClassName)}>
                    {this.props.label}
                </label>
            );
        }

        let groupClassName = {
            'form-group'   : !this.props.standalone,
            'form-group-lg': !this.props.standalone && this.props.bsSize === 'large',
            'form-group-sm': !this.props.standalone && this.props.bsSize === 'small',
            'has-feedback' : this.props.hasFeedback,
            'has-success'  : this.props.bsStyle === 'success',
            'has-warning'  : this.props.bsStyle === 'warning',
            'has-error'    : this.props.bsStyle === 'error'
        };

        return (
            <div className={classNames(groupClassName, this.props.groupClassName)}>
                {label}
                <div className={this.props.wrapperClassName}>
                    {this._renderChildren()}
                    {this._renderHelp()}
                </div>
            </div>
        );
    }

    _renderChildren() {
        return React.Children.map(this.props.children, child => {
            if (child.type !== Radio) {
                throw new Error('Only Radio component is allowed inside RadioGroup');
            }

            return React.cloneElement(child, {
                type            : 'radio',
                standalone      : true,
                checked         : this.state.value === child.props.value,
                name            : this.props.name,
                onChange        : this._onChange.bind(this),
                _registerInput  : this.registerInput.bind(this),
                _unregisterInput: this.unregisterInput.bind(this)
            });
        });
    }

    _renderHelp() {
        return this.props.help ? (
            <span className="help-block" key="help">
                {this.props.help}
            </span>
        ) : null;
    }

    _onChange(e) {
        if (!e.target) {
            return;
        }

        this.setState({
            value: e.target.value
        });

        this.props.onChange(e);
    }
}

RadioGroup.propTypes = {
    standalone      : PropTypes.bool,
    hasFeedback     : PropTypes.bool,
    bsSize (props) {
        if (props.standalone && props.bsSize !== undefined) {
            return new Error('bsSize will not be used when `standalone` is set.');
        }

        return PropTypes.oneOf([ 'small', 'medium', 'large' ])
            .apply(null, arguments);
    },
    bsStyle         : PropTypes.oneOf([ 'success', 'warning', 'error' ]),
    groupClassName  : PropTypes.string,
    wrapperClassName: PropTypes.string,
    labelClassName  : PropTypes.string,
    validationEvent : PropTypes.oneOf([
        'onChange'
    ]),
    validate        : PropTypes.oneOfType([
        PropTypes.func,
        PropTypes.string
    ]),
    errorHelp       : PropTypes.oneOfType([
        PropTypes.string,
        PropTypes.object
    ])
};

RadioGroup.defaultProps = {
    standalone     : false,
    validationEvent: 'onChange',
    onChange       : () => {}
};
