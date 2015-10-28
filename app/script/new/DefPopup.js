import React from "react";
import Draggable from "react-draggable";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import ExampleView from "./ExampleView";
import DiscussionsView from "./DiscussionsView";

export default class DefPopup extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			examplesGeneration: -1,
		};
	}

	componentWillReceiveProps(nextProps) {
		this.setState({examplesGeneration: nextProps.examples.generation});
	}

	shouldComponentUpdate(nextProps, nextState) {
		return nextProps.def !== this.props.def ||
			nextProps.highlightedDef !== this.props.highlightedDef ||
			nextProps.discussions !== this.props.discussions ||
			nextState.examplesGeneration !== this.state.examplesGeneration;
	}

	render() {
		let def = this.props.def;
		return (
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={(event) => {
								event.preventDefault();
								Dispatcher.dispatch(new DefActions.GoToDef(def.URL));
							}}>Go to definition</a>
							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</header>

						<section className="docHTML">
							<div className="header">
								<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
							</div>
							<section className="doc" dangerouslySetInnerHTML={def.Data && def.Data.DocHTML} />
						</section>

						<ExampleView defURL={def.URL} examples={this.props.examples} highlightedDef={this.props.highlightedDef} />

						{this.props.discussions && <DiscussionsView defURL={def.URL} discussions={this.props.discussions.slice(0, 4)} />}
					</div>
				</div>
			</Draggable>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
	discussions: React.PropTypes.array,
};
